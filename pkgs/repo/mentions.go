package repo

import (
	"html"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

// Mention is one @path or @path(start-end) token in a prompt (1-based line range, inclusive).
type Mention struct {
	Path      string
	StartLine int
	EndLine   int
	HasRange  bool
	RawStart  int
	RawEnd    int
}

var (
	dataPathAttrRe = regexp.MustCompile(`(?i)\bdata-path="([^"]+)"`)
	// Matches a full opening tag that declares data-path (chip or embed).
	dataPathTagRe = regexp.MustCompile(`(?i)<[^>]*\bdata-path="([^"]+)"[^>]*>`)
	lineStartRe   = regexp.MustCompile(`(?i)\bdata-line-start="(\d+)"`)
	lineEndRe     = regexp.MustCompile(`(?i)\bdata-line-end="(\d+)"`)
)

// ParseFileMentions extracts @-mentions. Paths may not contain whitespace; range uses (start-end).
// TipTap stores chips as HTML with data-path; those are preferred over a naive scan of raw HTML
// (which would otherwise treat "@x.go</span>" as a path).
func ParseFileMentions(s string) []Mention {
	slog.Debug("trace", "operation", "repo.ParseFileMentions")
	if looksLikePromptHTML(s) {
		if strings.Contains(strings.ToLower(s), "data-path=") {
			return parseHTMLDataPathMentions(s)
		}
		return parsePlainFileMentions(stripHTMLTags(s))
	}
	return parsePlainFileMentions(s)
}

func looksLikePromptHTML(s string) bool {
	return strings.Contains(s, "<") && strings.Contains(s, ">")
}

func parseHTMLDataPathMentions(s string) []Mention {
	var out []Mention
	for _, loc := range dataPathTagRe.FindAllStringSubmatchIndex(s, -1) {
		// loc: full start/end, path group start/end
		if len(loc) < 4 {
			continue
		}
		tag := s[loc[0]:loc[1]]
		path := html.UnescapeString(s[loc[2]:loc[3]])
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		m := Mention{
			Path:     path,
			RawStart: loc[0],
			RawEnd:   loc[1],
		}
		if start := lineStartRe.FindStringSubmatch(tag); len(start) == 2 {
			if end := lineEndRe.FindStringSubmatch(tag); len(end) == 2 {
				a, errA := strconv.Atoi(start[1])
				b, errB := strconv.Atoi(end[1])
				if errA == nil && errB == nil && a >= 1 && b >= a {
					m.StartLine = a
					m.EndLine = b
					m.HasRange = true
				}
			}
		}
		out = append(out, m)
	}
	if len(out) > 0 {
		return out
	}
	// data-path present but no tags matched (malformed) — fall back.
	_ = dataPathAttrRe
	return parsePlainFileMentions(stripHTMLTags(s))
}

func stripHTMLTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inTag := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' && inTag {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteByte(c)
		}
	}
	return html.UnescapeString(b.String())
}

func parsePlainFileMentions(s string) []Mention {
	var out []Mention
	i := 0
outer:
	for i < len(s) {
		j := strings.Index(s[i:], "@")
		if j < 0 {
			break
		}
		i += j
		rawStart := i
		i++
		pathStart := i
		for i < len(s) {
			c := s[i]
			if c == '(' {
				newI, mention, contOuter, restartFrom, restartOuter := handleMentionOpenParen(s, i, pathStart, rawStart)
				if restartOuter {
					i = restartFrom
					continue outer
				}
				i = newI
				if contOuter && mention != nil {
					out = append(out, *mention)
					continue outer
				}
				continue
			}
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '@' {
				break
			}
			i++
		}
		path := strings.TrimSpace(s[pathStart:i])
		if path != "" {
			out = append(out, Mention{
				Path:     path,
				RawStart: rawStart,
				RawEnd:   i,
			})
		}
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func isMentionDelimiter(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '@'
}
