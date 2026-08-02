package agentmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitstack"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ToolCreatePullRequest = "hamix.create_pull_request"
	maxPRTitleBytes       = 256
	maxPRBodyBytes        = 64 * 1024
)

// stackCLI is the gh stack runner (tests may replace).
var stackCLI gitstack.CLI = gitstack.New()

type createPullRequestTool struct{}

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (createPullRequestTool) Name() string { return ToolCreatePullRequest }

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (createPullRequestTool) Group() string { return GroupGit }

//funclogmeasure:skip category=hot-path reason="MCP tool metadata accessor."
func (createPullRequestTool) Description() string {
	return "Publish the worktree GitHub stack (gh stack submit) or open a single PR. Required to finish an open-pr run. Do not use Shell git push or gh pr create."
}

type createPullRequestInput struct {
	Title string `json:"title" jsonschema:"pull request title for the current (root) layer"`
	Body  string `json:"body" jsonschema:"pull request body in markdown for the current layer"`
	Base  string `json:"base,omitempty" jsonschema:"optional base branch override for fallback single-PR path"`
}

type createPullRequestOutput struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url"`
	Number int    `json:"number,omitempty"`
	Title  string `json:"title"`
	Base   string `json:"base,omitempty"`
	Head   string `json:"head,omitempty"`
}

//funclogmeasure:skip category=hot-path reason="MCP SDK registration; business logic is in runCreatePullRequest."
func (t createPullRequestTool) Register(server *mcp.Server, sess *Session) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Description(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, in createPullRequestInput) (*mcp.CallToolResult, createPullRequestOutput, error) {
		out, err := runCreatePullRequest(ctx, sess, in)
		if err != nil {
			return toolErr(err), createPullRequestOutput{}, nil
		}
		return nil, out, nil
	})
}

//funclogmeasure:skip category=hot-path reason="Git + gh I/O; callers surface errors to the agent."
func runCreatePullRequest(ctx context.Context, sess *Session, in createPullRequestInput) (createPullRequestOutput, error) {
	if sess.Phase != PhaseExecute {
		return createPullRequestOutput{}, fmt.Errorf("phase is %q; %s requires execute", sess.Phase, ToolCreatePullRequest)
	}
	workDir := strings.TrimSpace(sess.WorkingDir)
	if workDir == "" {
		return createPullRequestOutput{}, fmt.Errorf("working_dir is empty")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return createPullRequestOutput{}, fmt.Errorf("title is required")
	}
	if len(title) > maxPRTitleBytes {
		return createPullRequestOutput{}, fmt.Errorf("title too long")
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return createPullRequestOutput{}, fmt.Errorf("body is required")
	}
	if len(body) > maxPRBodyBytes {
		return createPullRequestOutput{}, fmt.Errorf("body too long")
	}
	base := strings.TrimSpace(in.Base)

	headOut, err := gitRun(ctx, workDir, "branch", "--show-current")
	if err != nil {
		return createPullRequestOutput{}, fmt.Errorf("resolve current branch: %w", err)
	}
	head := strings.TrimSpace(headOut)
	if head == "" {
		return createPullRequestOutput{}, fmt.Errorf("detached HEAD; checkout a branch before opening a PR")
	}

	// Prefer stacked publish (ADR-0097). Fall back to single gh pr create when
	// local stack tracking is unavailable.
	_ = stackCLI.Rebase(ctx, workDir)
	if _, submitErr := stackCLI.Submit(ctx, workDir); submitErr == nil {
		layers, layerErr := collectStackPRLayers(ctx, workDir)
		if layerErr != nil {
			return createPullRequestOutput{}, layerErr
		}
		url, number, prBase, prHead := pickLayer(layers, head)
		if url == "" {
			// Submit succeeded but view lag — resolve current branch PR.
			url, number, prBase, prHead, err = viewPR(ctx, workDir, head)
			if err != nil || url == "" {
				return createPullRequestOutput{}, fmt.Errorf("stack submit succeeded but no PR URL for %q", head)
			}
		}
		if prBase != "" {
			base = prBase
		}
		if prHead != "" {
			head = prHead
		}
		// Best-effort title/body edit on the current layer PR.
		_, _ = ghRun(ctx, workDir, "pr", "edit", head, "--title", title, "--body", body)
		rep := sidecar.PullRequestReport{
			SchemaVersion: sidecar.CurrentSchemaVersion,
			URL:           url,
			Number:        number,
			Title:         title,
			Base:          base,
			Head:          head,
			Layers:        layers,
		}
		if err := sidecar.WritePullRequestReport(sess.ReportDir, sess.CycleID, rep); err != nil {
			return createPullRequestOutput{}, err
		}
		if err := writeReceipt(sess, ToolCreatePullRequest, PhaseExecute, sidecar.PullRequestSubmitReceiptPath(sess.ReportDir, sess.CycleID)); err != nil {
			return createPullRequestOutput{}, err
		}
		return createPullRequestOutput{
			OK: true, URL: url, Number: number, Title: title, Base: base, Head: head,
		}, nil
	}

	pushOut, err := gitRun(ctx, workDir, "push", "-u", "origin", "HEAD")
	if err != nil {
		detail := strings.TrimSpace(pushOut)
		if detail == "" {
			detail = err.Error()
		}
		return createPullRequestOutput{}, fmt.Errorf("git push failed: %s", detail)
	}

	args := []string{"pr", "create", "--title", title, "--body", body, "--head", head}
	if base != "" {
		args = append(args, "--base", base)
	}
	prOut, err := ghRun(ctx, workDir, args...)
	if err != nil {
		detail := strings.TrimSpace(prOut)
		if detail == "" {
			detail = err.Error()
		}
		return createPullRequestOutput{}, fmt.Errorf("gh pr create failed: %s", detail)
	}

	url := firstHTTPURLLine(prOut)
	number := 0
	if u, n, b, h, viewErr := viewPR(ctx, workDir, head); viewErr == nil {
		if u != "" {
			url = u
		}
		number = n
		if base == "" {
			base = b
		}
		if h != "" {
			head = h
		}
	}
	if url == "" {
		return createPullRequestOutput{}, fmt.Errorf("gh pr create returned no URL: %q", strings.TrimSpace(prOut))
	}

	rep := sidecar.PullRequestReport{
		SchemaVersion: sidecar.CurrentSchemaVersion,
		URL:           url,
		Number:        number,
		Title:         title,
		Base:          base,
		Head:          head,
		Layers: []sidecar.PullRequestLayer{{
			Head: head, URL: url, Number: number, Base: base, Title: title,
		}},
	}
	if err := sidecar.WritePullRequestReport(sess.ReportDir, sess.CycleID, rep); err != nil {
		return createPullRequestOutput{}, err
	}
	if err := writeReceipt(sess, ToolCreatePullRequest, PhaseExecute, sidecar.PullRequestSubmitReceiptPath(sess.ReportDir, sess.CycleID)); err != nil {
		return createPullRequestOutput{}, err
	}
	return createPullRequestOutput{
		OK: true, URL: url, Number: number, Title: title, Base: base, Head: head,
	}, nil
}

func collectStackPRLayers(ctx context.Context, workDir string) ([]sidecar.PullRequestLayer, error) {
	raw, err := stackCLI.ViewJSON(ctx, workDir)
	if err != nil {
		return nil, fmt.Errorf("gh stack view: %w", err)
	}
	var view struct {
		Branches []struct {
			Name string `json:"name"`
		} `json:"branches"`
	}
	if err := json.Unmarshal([]byte(raw), &view); err != nil {
		return nil, fmt.Errorf("parse stack view: %w", err)
	}
	var layers []sidecar.PullRequestLayer
	for _, b := range view.Branches {
		name := strings.TrimSpace(b.Name)
		if name == "" {
			continue
		}
		url, number, base, head, viewErr := viewPR(ctx, workDir, name)
		if viewErr != nil || url == "" {
			continue
		}
		if head == "" {
			head = name
		}
		layers = append(layers, sidecar.PullRequestLayer{
			Head: head, URL: url, Number: number, Base: base,
		})
	}
	return layers, nil
}

func pickLayer(layers []sidecar.PullRequestLayer, head string) (url string, number int, base, prHead string) {
	for _, l := range layers {
		if l.Head == head {
			return l.URL, l.Number, l.Base, l.Head
		}
	}
	if len(layers) > 0 {
		l := layers[0]
		return l.URL, l.Number, l.Base, l.Head
	}
	return "", 0, "", ""
}

func viewPR(ctx context.Context, workDir, head string) (url string, number int, base, prHead string, err error) {
	viewOut, viewErr := ghRun(ctx, workDir, "pr", "view", head, "--json", "url,number,baseRefName,headRefName")
	if viewErr != nil {
		return "", 0, "", "", viewErr
	}
	var view struct {
		URL         string `json:"url"`
		Number      int    `json:"number"`
		BaseRefName string `json:"baseRefName"`
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal([]byte(viewOut), &view); err != nil {
		return "", 0, "", "", err
	}
	return strings.TrimSpace(view.URL), view.Number, view.BaseRefName, strings.TrimSpace(view.HeadRefName), nil
}

//funclogmeasure:skip category=hot-path reason="Thin gh exec wrapper."
func ghRun(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func firstHTTPURLLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "://") {
			return line
		}
	}
	return ""
}
