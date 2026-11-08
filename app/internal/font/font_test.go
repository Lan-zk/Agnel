package font

import "testing"

func TestEmbeddedFontsPresentAndRegister(t *testing.T) {
	if len(regularTTF) < 100000 {
		t.Fatalf("regularTTF too small: %d bytes", len(regularTTF))
	}
	if len(boldTTF) < 100000 {
		t.Fatalf("boldTTF too small: %d bytes", len(boldTTF))
	}
	if len(italicTTF) < 50000 {
		t.Fatalf("italicTTF too small: %d bytes", len(italicTTF))
	}
	if got := Register(); got == "" {
		t.Fatal("Register returned empty family")
	}
}
