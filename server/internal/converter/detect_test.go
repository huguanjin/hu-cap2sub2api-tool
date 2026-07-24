package converter

import "testing"

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		name string
		doc  interface{}
		want string
	}{
		{"array is sub2api", []interface{}{anyMap{"a": 1}}, "sub2api"},
		{"accounts field", anyMap{"accounts": []interface{}{}}, "sub2api"},
		{"platform+credentials", anyMap{"platform": "claude", "credentials": anyMap{}}, "sub2api"},
		{"cpa type codex", anyMap{"type": "codex"}, "cpa"},
		{"cpa type claude", anyMap{"type": "claude"}, "cpa"},
		{"cpa type antigravity", anyMap{"type": "antigravity"}, "cpa"},
		{"cpa type gemini", anyMap{"type": "gemini"}, "cpa"},
		{"unknown explicit type", anyMap{"type": "something-else"}, "unknown"},
		{"access+id token no type", anyMap{"access_token": "x", "id_token": "y"}, "cpa"},
		{"access token only, no id_token", anyMap{"access_token": "x"}, "unknown"},
		{"token.access_token no type", anyMap{"token": anyMap{"access_token": "x"}}, "cpa"},
		{"token.accessToken camelCase", anyMap{"token": anyMap{"accessToken": "x"}}, "cpa"},
		{"empty object", anyMap{}, "unknown"},
		{"non object non array", "just a string", "unknown"},
		{"nil", nil, "unknown"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DetectFormat(c.doc); got != c.want {
				t.Errorf("DetectFormat(%#v) = %q, want %q", c.doc, got, c.want)
			}
		})
	}
}
