package domain

import "testing"

func TestApplyPreservesContentOutsideSelection(t *testing.T) {
	got, err := Apply("before\nTARGET\nafter", Selection{Start: 7, End: 13}, "TARGET", "replacement")
	if err != nil || got != "before\nreplacement\nafter" {
		t.Fatalf("Apply() = %q, %v", got, err)
	}
}

func TestUTF16OffsetsMapToUTF8Boundaries(t *testing.T) {
	source := "A—🙂target"
	start, ok := UTF16OffsetToByte(source, 4)
	if !ok || source[start:] != "target" {
		t.Fatalf("offset mapped to %d, %t (%q)", start, ok, source[start:])
	}
	if _, ok := UTF16OffsetToByte(source, 3); ok {
		t.Fatal("accepted an offset inside a surrogate pair")
	}
}
