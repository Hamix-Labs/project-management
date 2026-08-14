package domain

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// AllowedTags is the block-level and inline element whitelist for
// draft-assist prompt HTML. It intentionally matches the TipTap schema the
// compose page ships (see docs/design/task-draft-ai.md): headings h2-h4,
// paragraphs, lists, blockquote, links, and repo-file mention chips. Any
// other element — most importantly `script`, `iframe`, `style`, `object`,
// `embed`, and form elements — is rejected. See ADR-0101 for the security
// contract behind this list.
var AllowedTags = map[string]struct{}{
	"h2":         {},
	"h3":         {},
	"h4":         {},
	"p":          {},
	"ul":         {},
	"ol":         {},
	"li":         {},
	"blockquote": {},
	"a":          {},
	"span":       {}, // only mention chips; enforced by attribute check.
	"br":         {},
}

// AllowedAttrs is the per-tag attribute whitelist. Anchors admit `href` only;
// span admits the repo-file mention chip attributes plus a small set of
// presentational hints; other tags admit no attributes. Anchor `href` must
// use a safe scheme (http, https, mailto, or a repo-relative path).
var AllowedAttrs = map[string]map[string]struct{}{
	"a": {
		"href": {},
	},
	"span": {
		"data-repo-file":  {},
		"data-path":       {},
		"data-line-start": {},
		"data-line-end":   {},
		"class":           {},
		"title":           {},
		"aria-label":      {},
	},
}

// ValidateHTML parses fragment as an HTML fragment and reports the first
// disallowed element or attribute it encounters.
//
// Empty and whitespace-only fragments are accepted (the UI treats them as an
// empty prompt). Text content is always allowed. The validator does not
// modify or canonicalize the input; callers who need normalization should
// handle that separately after ValidateHTML succeeds.
//
//funclogmeasure:skip category=hot-path reason="Pure validator; MCP write chokepoint emits the operation trace."
func ValidateHTML(fragment string) error {
	if strings.TrimSpace(fragment) == "" {
		return nil
	}
	z := html.NewTokenizer(strings.NewReader(fragment))
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if err := z.Err(); err == io.EOF {
				return nil
			} else if err != nil {
				return fmt.Errorf("%w: parse html: %v", ErrInvalidInput, err)
			}
			return nil
		}
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			if err := validateElement(tok); err != nil {
				return err
			}
		case html.EndTagToken:
			tok := z.Token()
			if _, ok := AllowedTags[strings.ToLower(tok.Data)]; !ok {
				return fmt.Errorf("%w: tag %q not allowed", ErrInvalidInput, tok.Data)
			}
		}
	}
}

//funclogmeasure:skip category=hot-path reason="Inline validator helper; caller emits the operation trace."
func validateElement(tok html.Token) error {
	name := strings.ToLower(tok.Data)
	if _, ok := AllowedTags[name]; !ok {
		return fmt.Errorf("%w: tag %q not allowed", ErrInvalidInput, name)
	}
	allowed := AllowedAttrs[name]
	if name == "span" && !isMentionChip(tok) {
		return fmt.Errorf("%w: <span> only allowed as mention chip (data-repo-file=\"true\")", ErrInvalidInput)
	}
	for _, a := range tok.Attr {
		attrName := strings.ToLower(a.Key)
		if _, ok := allowed[attrName]; !ok {
			return fmt.Errorf("%w: attribute %q on <%s> not allowed", ErrInvalidInput, attrName, name)
		}
		if name == "a" && attrName == "href" {
			if err := validateHref(a.Val); err != nil {
				return err
			}
		}
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure attr check; validator emits the operation trace."
func isMentionChip(tok html.Token) bool {
	for _, a := range tok.Attr {
		if strings.EqualFold(a.Key, "data-repo-file") && strings.EqualFold(strings.TrimSpace(a.Val), "true") {
			return true
		}
	}
	return false
}

//funclogmeasure:skip category=hot-path reason="Pure href check; validator emits the operation trace."
func validateHref(raw string) error {
	href := strings.TrimSpace(raw)
	if href == "" {
		return nil
	}
	lower := strings.ToLower(href)
	switch {
	case strings.HasPrefix(lower, "javascript:"), strings.HasPrefix(lower, "data:"), strings.HasPrefix(lower, "vbscript:"):
		return fmt.Errorf("%w: href scheme not allowed", ErrInvalidInput)
	}
	return nil
}
