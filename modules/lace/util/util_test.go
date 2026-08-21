package util

import (
	"testing"
	"time"
)

// TestISO pins the wire shape GraphQL string-time fields depend on: RFC3339
// in UTC, regardless of the input's zone.
func TestISO(t *testing.T) {
	loc := time.FixedZone("IST", 5*3600+1800)
	in := time.Date(2026, 8, 18, 12, 30, 0, 0, loc)
	if got, want := ISO(in), "2026-08-18T07:00:00Z"; got != want {
		t.Errorf("ISO = %q, want %q", got, want)
	}
}

// TestISOPtr — nil in → nil out; non-nil follows ISO.
func TestISOPtr(t *testing.T) {
	if got := ISOPtr(nil); got != nil {
		t.Errorf("ISOPtr(nil) = %v, want nil", *got)
	}
	in := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := ISOPtr(&in); got == nil || *got != "2026-01-02T03:04:05Z" {
		t.Errorf("ISOPtr = %v, want 2026-01-02T03:04:05Z", got)
	}
}

// TestStrPtr — the empty string maps to ABSENT (nil), never a pointer to "":
// nullable GraphQL fields must distinguish unset from empty.
func TestStrPtr(t *testing.T) {
	if got := StrPtr(""); got != nil {
		t.Errorf("StrPtr(\"\") = %q, want nil", *got)
	}
	if got := StrPtr("x"); got == nil || *got != "x" {
		t.Errorf("StrPtr(\"x\") = %v, want pointer to \"x\"", got)
	}
}

// TestDeref — nil takes the default; non-nil wins.
func TestDeref(t *testing.T) {
	if got := Deref[int](nil, 7); got != 7 {
		t.Errorf("Deref(nil, 7) = %d, want 7", got)
	}
	v := 3
	if got := Deref(&v, 7); got != 3 {
		t.Errorf("Deref(&3, 7) = %d, want 3", got)
	}
}
