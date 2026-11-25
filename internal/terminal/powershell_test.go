package terminal

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBoundedOutputAtLimit(t *testing.T) {
	var out boundedOutput
	full := strings.Repeat("a", maxOutputBytes)
	n, err := out.Write([]byte(full))
	if n != len(full) || err != nil {
		t.Fatalf("Write returned (%d, %v), want (%d, nil)", n, err, len(full))
	}
	if out.String() != full {
		t.Fatal("output at the limit should be preserved without a truncation marker")
	}

	for _, chunk := range []string{"tail", strings.Repeat("b", 2*maxOutputBytes)} {
		n, err = out.Write([]byte(chunk))
		if n != len(chunk) || err != nil {
			t.Fatalf("overflow Write returned (%d, %v), want (%d, nil)", n, err, len(chunk))
		}
	}
	if out.String() != full+outputTruncatedMarker {
		t.Fatal("overflow should preserve the retained prefix and append one marker")
	}
}

func TestBoundedOutputWhenLimitSplitsUTF8(t *testing.T) {
	var out boundedOutput
	prefix := strings.Repeat("a", maxOutputBytes-1)
	input := prefix + "é"
	n, err := out.Write([]byte(input))
	if n != len(input) || err != nil {
		t.Fatalf("Write returned (%d, %v), want (%d, nil)", n, err, len(input))
	}
	got := out.String()
	if !utf8.ValidString(got) || got != prefix+outputTruncatedMarker {
		t.Fatal("truncating inside a UTF-8 sequence should preserve a valid output prefix")
	}
}
