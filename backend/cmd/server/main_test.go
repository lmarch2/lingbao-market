package main

import "testing"

func TestParseCleanupTime(t *testing.T) {
	t.Parallel()

	t.Run("midnight", func(t *testing.T) {
		hour, minute, err := parseCleanupTime("00:00")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if hour != 0 || minute != 0 {
			t.Fatalf("expected 0:0, got %d:%d", hour, minute)
		}
	})

	t.Run("other", func(t *testing.T) {
		hour, minute, err := parseCleanupTime("23:59")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if hour != 23 || minute != 59 {
			t.Fatalf("expected 23:59, got %d:%d", hour, minute)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, _, err := parseCleanupTime("24:00")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
