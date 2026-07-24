package converter

import "testing"

func TestBuildMergedSub2ApiDocument_Empty(t *testing.T) {
	if got := BuildMergedSub2ApiDocument(nil, ConvertOptions{}); got != nil {
		t.Fatalf("expected nil for empty input, got %#v", got)
	}
}

func TestBuildMergedSub2ApiDocument_Single(t *testing.T) {
	results := []*CPAConversionResult{
		{Account: anyMap{"name": "a"}, OutputFileName: "a.sub2api.json"},
	}
	merged := BuildMergedSub2ApiDocument(results, ConvertOptions{})
	if merged.OutputFileName != "a.sub2api.json" {
		t.Errorf("OutputFileName = %q", merged.OutputFileName)
	}
	accounts, ok := merged.Document["accounts"].([]interface{})
	if !ok || len(accounts) != 1 {
		t.Fatalf("accounts = %#v", merged.Document["accounts"])
	}
	if merged.Count != 1 {
		t.Errorf("Count = %d, want 1", merged.Count)
	}
}

func TestBuildMergedSub2ApiDocument_Multiple(t *testing.T) {
	results := []*CPAConversionResult{
		{Account: anyMap{"name": "a"}, OutputFileName: "a.sub2api.json"},
		{Account: anyMap{"name": "b"}, OutputFileName: "b.sub2api.json"},
	}
	merged := BuildMergedSub2ApiDocument(results, ConvertOptions{})
	if merged.OutputFileName != "merged-accounts.sub2api.json" {
		t.Errorf("OutputFileName = %q", merged.OutputFileName)
	}
	accounts := merged.Document["accounts"].([]interface{})
	if len(accounts) != 2 {
		t.Errorf("accounts len = %d, want 2", len(accounts))
	}
	if merged.Count != 2 {
		t.Errorf("Count = %d, want 2", merged.Count)
	}
	if _, ok := merged.Document["exported_at"].(string); !ok {
		t.Error("exported_at should be a non-empty string")
	}
	if proxies, ok := merged.Document["proxies"].([]interface{}); !ok || len(proxies) != 0 {
		t.Errorf("proxies should be an empty array, got %#v", merged.Document["proxies"])
	}
}
