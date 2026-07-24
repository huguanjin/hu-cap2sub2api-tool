package converter

import "testing"

func TestSanitizeBaseName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"simple with ext", "simple.json", "simple"},
		{"mixed case spaces", "My File Name.JSON", "my-file-name"},
		{"unsafe chars collapse", "weird:name?*.txt", "weird-name"},
		{"whitespace runs", "   spaced   name  .json", "spaced-name"},
		{"empty string", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SanitizeBaseName(c.input); got != c.want {
				t.Errorf("SanitizeBaseName(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestBuildSub2ApiOutputFileName(t *testing.T) {
	if got := BuildSub2ApiOutputFileName("", ""); got != "converted-account.sub2api.json" {
		t.Errorf("both empty = %q", got)
	}
	if got := BuildSub2ApiOutputFileName("folder/sub/My Export.json", ""); got != "my-export.sub2api.json" {
		t.Errorf("sourceName with path = %q", got)
	}
	if got := BuildSub2ApiOutputFileName("", "admin@example"); got != "admin@example.sub2api.json" {
		t.Errorf("email fallback = %q", got)
	}
	if got := BuildSub2ApiOutputFileName("src.json", "email@x"); got != "src.sub2api.json" {
		t.Errorf("sourceName should take priority over email, got %q", got)
	}
}

func TestBuildCPAOutputFileName(t *testing.T) {
	if got := BuildCPAOutputFileName("MyAccount", "user@corp", "codex"); got != "user@corp.codex.cpa.json" {
		t.Errorf("email priority = %q", got)
	}
	if got := BuildCPAOutputFileName("", "", "claude"); got != "claude.cpa.json" {
		t.Errorf("base==typeBase collapse = %q", got)
	}
	if got := BuildCPAOutputFileName("", "", ""); got != "converted-account.account.cpa.json" {
		t.Errorf("all empty fallback = %q", got)
	}
}

func TestBuildMergedSub2ApiFileName(t *testing.T) {
	single := []*CPAConversionResult{{OutputFileName: "a.sub2api.json"}}
	if got := BuildMergedSub2ApiFileName(single); got != "a.sub2api.json" {
		t.Errorf("single record = %q", got)
	}

	multi := []*CPAConversionResult{{OutputFileName: "a.sub2api.json"}, {OutputFileName: "b.sub2api.json"}}
	if got := BuildMergedSub2ApiFileName(multi); got != "merged-accounts.sub2api.json" {
		t.Errorf("multi record = %q", got)
	}
}
