package converter

import "testing"

func TestStripUnavailable_ScalarValues(t *testing.T) {
	if _, ok := StripUnavailable(nil); ok {
		t.Error("nil should strip away")
	}
	if _, ok := StripUnavailable(""); ok {
		t.Error("empty string should strip away")
	}
	if v, ok := StripUnavailable("x"); !ok || v != "x" {
		t.Errorf("non-empty string should survive, got %v, %v", v, ok)
	}
	if v, ok := StripUnavailable(float64(0)); !ok || v != float64(0) {
		t.Errorf("zero number should survive (only nil/empty-string/empty-object strip), got %v, %v", v, ok)
	}
	if v, ok := StripUnavailable(false); !ok || v != false {
		t.Errorf("false bool should survive, got %v, %v", v, ok)
	}
}

func TestStripUnavailable_ArraysAlwaysKept(t *testing.T) {
	v, ok := StripUnavailable([]interface{}{})
	if !ok {
		t.Fatal("empty array should be kept, not stripped")
	}
	arr, isArr := v.([]interface{})
	if !isArr || len(arr) != 0 {
		t.Fatalf("expected empty array, got %#v", v)
	}

	v2, ok := StripUnavailable([]interface{}{"", nil, "x"})
	if !ok {
		t.Fatal("array with at least one surviving element should be kept")
	}
	arr2 := v2.([]interface{})
	if len(arr2) != 1 || arr2[0] != "x" {
		t.Errorf("expected [\"x\"], got %#v", arr2)
	}
}

func TestStripUnavailable_EmptyObjectDropped(t *testing.T) {
	if _, ok := StripUnavailable(anyMap{}); ok {
		t.Error("empty object should strip away")
	}
	if _, ok := StripUnavailable(anyMap{"a": ""}); ok {
		t.Error("object whose every field strips away should itself strip away")
	}
}

func TestStripUnavailable_PartialObjectKept(t *testing.T) {
	v, ok := StripUnavailable(anyMap{"a": "", "b": "x"})
	if !ok {
		t.Fatal("object with at least one surviving field should be kept")
	}
	m := v.(anyMap)
	if _, exists := m["a"]; exists {
		t.Error("empty field a should have been dropped")
	}
	if m["b"] != "x" {
		t.Errorf("b = %v, want x", m["b"])
	}
}

func TestStripUnavailable_NestedObjectAllEmptyDropsOuter(t *testing.T) {
	if _, ok := StripUnavailable(anyMap{"outer": anyMap{"inner": ""}}); ok {
		t.Error("nested object stripping to empty should cause outer object to also strip away")
	}
}

func TestStripMap_ReturnsEmptyMapNotNil(t *testing.T) {
	m := stripMap(anyMap{"a": ""})
	if m == nil {
		t.Fatal("stripMap should never return nil")
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %#v", m)
	}
}
