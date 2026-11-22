package provider

import "testing"

func TestNormalizeEmailForwardAddress(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"hello":               "hello",
		"hello@example.com":   "hello",
		" a_b+c@example.com ": "a_b+c",
		"":                    "",
	} {
		if got := normalizeEmailForwardAddress(in); got != want {
			t.Errorf("normalizeEmailForwardAddress(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEmailForwardMatches(t *testing.T) {
	t.Parallel()

	if !emailForwardMatches("Hello@example.com", "Me@example.org", "hello", "me@example.org") {
		t.Error("expected case-insensitive match on local part and destination")
	}
	if emailForwardMatches("hello@example.com", "me@example.org", "hello", "other@example.org") {
		t.Error("expected no match for a different destination")
	}
	if emailForwardMatches("hello@example.com", "me@example.org", "bye", "me@example.org") {
		t.Error("expected no match for a different address")
	}
}
