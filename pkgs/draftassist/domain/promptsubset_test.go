package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

func TestValidateHTML_allowed(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{"empty", ""},
		{"whitespace", "   \n\t  "},
		{"plain text", "hello world"},
		{"paragraph", "<p>hello</p>"},
		{"heading h2", "<h2>Section</h2>"},
		{"heading h3", "<h3>Section</h3>"},
		{"heading h4", "<h4>Section</h4>"},
		{"list ul", "<ul><li>one</li><li>two</li></ul>"},
		{"list ol", "<ol><li>first</li></ol>"},
		{"blockquote", "<blockquote><p>quoted</p></blockquote>"},
		{"anchor http", `<a href="https://example.com">link</a>`},
		{"anchor mailto", `<a href="mailto:you@example.com">mail</a>`},
		{"anchor relative", `<a href="/docs/api.md">docs</a>`},
		{"br in paragraph", `<p>first<br/>second</p>`},
		{"mention chip", `<p>See <span data-repo-file="true" data-path="pkgs/x" class="repo-file-chip" title="@pkgs/x" aria-label="File reference: @pkgs/x">@pkgs/x</span>.</p>`},
		{"mention chip with range", `<span data-repo-file="true" data-path="a" data-line-start="1" data-line-end="10">@a(1-10)</span>`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := domain.ValidateHTML(tc.html); err != nil {
				t.Fatalf("ValidateHTML(%q) unexpected error: %v", tc.html, err)
			}
		})
	}
}

func TestValidateHTML_rejects(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{"script tag", `<script>alert(1)</script>`},
		{"script inside p", `<p>hi <script>alert(1)</script></p>`},
		{"iframe", `<iframe src="https://evil"></iframe>`},
		{"style", `<style>body { display: none; }</style>`},
		{"object", `<object data="x"></object>`},
		{"embed", `<embed src="x"/>`},
		{"form", `<form><input/></form>`},
		{"img tag", `<img src="x"/>`},
		{"h1 heading", `<h1>too big</h1>`},
		{"span without mention marker", `<span class="foo">bare</span>`},
		{"span with wrong marker value", `<span data-repo-file="false">no</span>`},
		{"onclick on p", `<p onclick="steal()">hi</p>`},
		{"style attr on p", `<p style="color:red">hi</p>`},
		{"javascript href", `<a href="javascript:alert(1)">x</a>`},
		{"data: href", `<a href="data:text/html,evil">x</a>`},
		{"disallowed target attr on a", `<a href="/" target="_blank">x</a>`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateHTML(tc.html)
			if err == nil {
				t.Fatalf("ValidateHTML(%q) unexpectedly accepted", tc.html)
			}
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("ValidateHTML(%q) err=%v; want ErrInvalidInput", tc.html, err)
			}
		})
	}
}

func TestValidateHTML_endTagCheck(t *testing.T) {
	if err := domain.ValidateHTML(`<p>hi</p></script>`); err == nil {
		t.Fatal("expected rejection for stray </script>")
	} else if !strings.Contains(err.Error(), "script") {
		t.Fatalf("expected script in err; got %v", err)
	}
}
