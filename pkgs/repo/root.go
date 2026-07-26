package repo

import (
	"errors"
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Empty q lists the first N files (walk order) for @-mention browse; non-empty q caps matches for performance.
	maxSearchResultsBrowse = 250
	maxSearchResultsFilter = 100
	maxFileReadBytes       = 32 << 20 // 32 MiB upper bound for line counting
)

// Root is a validated absolute directory used for repo-relative paths.
type Root struct {
	abs   string
	canon string
}

// OpenRoot returns a Root for dir, or ErrInvalidInput if missing or not a directory.
func OpenRoot(dir string) (*Root, error) {
	slog.Debug("trace", "operation", "repo.OpenRoot")
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("%w: repo root is empty", taskcoredomain.ErrInvalidInput)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: repo root: %v", taskcoredomain.ErrInvalidInput, err)
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("%w: repo root: %v", taskcoredomain.ErrInvalidInput, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%w: repo root is not a directory", taskcoredomain.ErrInvalidInput)
	}
	canon, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("%w: repo root symlink resolution: %v", taskcoredomain.ErrInvalidInput, err)
	}
	return &Root{abs: abs, canon: filepath.Clean(canon)}, nil
}

// Abs returns the absolute root path.
func (r *Root) Abs() string {
	slog.Debug("trace", "operation", "repo.Root.Abs")
	return r.abs
}

// Ready verifies the root directory still exists and is a directory (for HTTP readiness when a repo root is configured via Settings).
func (r *Root) Ready() error {
	slog.Debug("trace", "operation", "repo.Root.Ready")
	if r == nil {
		return errors.New("repo: nil root")
	}
	fi, err := os.Stat(r.abs)
	if err != nil {
		return fmt.Errorf("repo root: %w", err)
	}
	if !fi.IsDir() {
		return errors.New("repo root is not a directory")
	}
	return nil
}

// Resolve returns an absolute path for a repo-relative path (slashes or OS separators).
func (r *Root) Resolve(rel string) (string, error) {
	slog.Debug("trace", "operation", "repo.Root.Resolve")
	rel = strings.TrimSpace(rel)
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return "", fmt.Errorf("%w: invalid path", taskcoredomain.ErrInvalidInput)
	}
	// Reject only when ".." appears as its own path component
	// (a parent-directory traversal). The previous broad
	// strings.Contains(rel, "..") check rejected legitimate filenames
	// like "foo..bar.go", "eslint..config.js", and "..gitkeep". The
	// downstream pathEscapesRoot(relOut) check after filepath.Rel is
	// the authoritative containment guard for any traversal that
	// survives Clean (e.g. "../outside" → "..", "a/../../b" → "../b").
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." {
			return "", fmt.Errorf("%w: invalid path", taskcoredomain.ErrInvalidInput)
		}
	}
	joined := filepath.Join(r.abs, filepath.FromSlash(rel))
	clean := filepath.Clean(joined)
	rootClean := filepath.Clean(r.abs)
	relOut, err := filepath.Rel(rootClean, clean)
	if err != nil || pathEscapesRoot(relOut) {
		return "", fmt.Errorf("%w: path escapes repo root", taskcoredomain.ErrInvalidInput)
	}
	targetCanonical, err := canonicalizePathForContainment(clean)
	if err != nil {
		return "", fmt.Errorf("%w: path canonicalization failed: %v", taskcoredomain.ErrInvalidInput, err)
	}
	canonicalRel, err := filepath.Rel(r.canon, targetCanonical)
	if err != nil || pathEscapesRoot(canonicalRel) {
		return "", fmt.Errorf("%w: path escapes repo root via symlink", taskcoredomain.ErrInvalidInput)
	}
	return clean, nil
}

// ValidatePromptMentions checks every parsed mention against the repo root.
func (r *Root) ValidatePromptMentions(prompt string) error {
	slog.Debug("trace", "operation", "repo.Root.ValidatePromptMentions")
	resolvedPaths := make(map[string]string)
	seenFiles := make(map[string]struct{})
	lineCounts := make(map[string]int)
	for _, m := range ParseFileMentions(prompt) {
		abs, ok := resolvedPaths[m.Path]
		if !ok {
			var err error
			abs, err = r.Resolve(m.Path)
			if err != nil {
				return wrapMention(m.mentionLabel(), err)
			}
			resolvedPaths[m.Path] = abs
		}
		if _, ok := seenFiles[abs]; !ok {
			fi, err := os.Stat(abs)
			if err != nil {
				if os.IsNotExist(err) {
					return wrapMentionMsg(m.mentionLabel(), "file does not exist")
				}
				return wrapMention(m.mentionLabel(), err)
			}
			if fi.IsDir() {
				return wrapMentionMsg(m.mentionLabel(), "path is a directory, not a file")
			}
			seenFiles[abs] = struct{}{}
		}
		if m.HasRange {
			if err := validateRangeBounds(m.StartLine, m.EndLine); err != nil {
				return wrapMention(m.mentionLabel(), err)
			}
			n, ok := lineCounts[abs]
			if !ok {
				lineCount, lineErr := LineCount(abs)
				if lineErr != nil {
					return wrapMention(m.mentionLabel(), lineErr)
				}
				n = lineCount
				lineCounts[abs] = n
			}
			if err := validateRangeWithLineCount(m.StartLine, m.EndLine, n); err != nil {
				return wrapMention(m.mentionLabel(), err)
			}
		}
	}
	return nil
}

// invalidInputPrefix is the text form of taskcoredomain.ErrInvalidInput as it
// appears when stringified via fmt.Errorf("%w: ...", taskcoredomain.ErrInvalidInput).
// Kept as a package var initialized once (rather than recomputed on every
// wrap) since hot prompts can carry many mentions.
var invalidInputPrefix = taskcoredomain.ErrInvalidInput.Error() + ": "

// wrapMention wraps cause with a single "tasks: invalid input: <label>: <reason>"
// prefix. If cause already carries the "tasks: invalid input: " prefix (it
// will whenever cause came from r.Resolve, validateRangeBounds, LineCount, or
// any other ErrInvalidInput-wrapped sink in this package), the duplicated
// prefix is stripped from cause's message so clients see exactly one prefix
// on the wire instead of the historical doubled "tasks: invalid input: mention
// @<path>: tasks: invalid input: <reason>". errors.Is(err, taskcoredomain.ErrInvalidInput)
// remains true via the %w on taskcoredomain.ErrInvalidInput, so all existing 400
// mappings in pkgs/tasks/handlerhttp.StoreErrorClientMessage
// continue to fire unchanged.
func wrapMention(label string, cause error) error {
	slog.Debug("trace", "operation", "repo.wrapMention")
	return wrapMentionMsg(label, strings.TrimPrefix(cause.Error(), invalidInputPrefix))
}

// wrapMentionMsg is the literal-message variant of wrapMention used when the
// reason is a hardcoded sentinel string ("file does not exist",
// "path is a directory, not a file") rather than an error value. Keeping a
// single shared format keeps the wire phrase identical across both code paths.
func wrapMentionMsg(label, reason string) error {
	slog.Debug("trace", "operation", "repo.wrapMentionMsg")
	return fmt.Errorf("%w: %s: %s", taskcoredomain.ErrInvalidInput, label, reason)
}

// mentionLabel formats the human-readable "mention @<path>" or
// "mention @<path>(<start>-<end>)" prefix used in error messages so range
// failures still pinpoint the offending line span (the substring clients in
// docs/api.md are documented to rely on for form UI highlighting).
func (m Mention) mentionLabel() string {
	slog.Debug("trace", "operation", "repo.Mention.mentionLabel")
	if m.HasRange {
		return fmt.Sprintf("mention @%s(%d-%d)", m.Path, m.StartLine, m.EndLine)
	}
	return "mention @" + m.Path
}
