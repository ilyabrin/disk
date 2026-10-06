package disk

import (
	"bytes"
	"strings"
	"testing"
)

const probeToken = "y0_AgAAAAREALSECRETTOKENxyz"

// logHeaders runs LogRequest at full verbosity and returns the header lines.
func logHeaders(t *testing.T, sanitize bool, headers map[string]string) []string {
	t.Helper()
	var buf bytes.Buffer
	cfg := DefaultLoggerConfig()
	cfg.Level = DEBUG
	cfg.Verbose = true
	cfg.SanitizeAuth = sanitize
	cfg.Structured = false
	cfg.Output = &buf

	NewLogger(cfg).LogRequest("GET", "https://example.invalid/disk", headers)

	var lines []string
	for _, line := range strings.Split(buf.String(), "\n") {
		if i := strings.Index(line, "Header: "); i >= 0 {
			lines = append(lines, line[i+len("Header: "):])
		}
	}
	return lines
}

func TestLogRequestShowsOrdinaryHeaders(t *testing.T) {
	lines := logHeaders(t, true, map[string]string{"Content-Type": "application/json"})
	if len(lines) != 1 || lines[0] != "Content-Type: application/json" {
		t.Errorf("an ordinary header should be logged as is, got %q", lines)
	}
}

func TestLogRequestMasksTheWholeCredential(t *testing.T) {
	lines := logHeaders(t, true, map[string]string{"Authorization": "OAuth " + probeToken})
	if len(lines) != 1 || lines[0] != "Authorization: OAuth ***" {
		t.Fatalf("want the scheme kept and the credential masked, got %q", lines)
	}
	// No fragment of the token may survive, not even a few characters.
	for _, n := range []int{2, 3, 4} {
		if strings.Contains(lines[0], probeToken[len(probeToken)-n:]) ||
			strings.Contains(lines[0], probeToken[:n]) {
			t.Errorf("part of the token leaked into %q", lines[0])
		}
	}
}

func TestLogRequestWithSanitizeAuthOffShowsEverything(t *testing.T) {
	lines := logHeaders(t, false, map[string]string{
		"Authorization": "OAuth " + probeToken,
		"Content-Type":  "application/json",
	})
	want := []string{
		"Authorization: OAuth " + probeToken,
		"Content-Type: application/json",
	}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("SanitizeAuth=false should log values unchanged\n got: %q\nwant: %q", lines, want)
	}
}

func TestLogRequestSortsHeaders(t *testing.T) {
	headers := map[string]string{"Zeta": "1", "Alpha": "2", "Mid": "3", "Beta": "4"}
	want := "Alpha: 2|Beta: 4|Mid: 3|Zeta: 1"
	// Map iteration order is random, so repeat to catch an unsorted loop.
	for i := 0; i < 20; i++ {
		if got := strings.Join(logHeaders(t, true, headers), "|"); got != want {
			t.Fatalf("headers should be logged in name order\n got: %s\nwant: %s", got, want)
		}
	}
}

func TestLogRequestOmitsHeadersWhenNotVerbose(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultLoggerConfig()
	cfg.Level = DEBUG
	cfg.Verbose = false
	cfg.Output = &buf

	NewLogger(cfg).LogRequest("GET", "https://example.invalid/disk",
		map[string]string{"Authorization": "OAuth " + probeToken})

	if strings.Contains(buf.String(), "Header") || strings.Contains(buf.String(), probeToken) {
		t.Errorf("headers must not be logged without Verbose, got %q", buf.String())
	}
}

func TestSanitizeValue(t *testing.T) {
	cases := []struct {
		name, key, value, want string
	}{
		{"oauth keeps scheme", "Authorization", "OAuth " + probeToken, "OAuth ***"},
		{"bearer keeps scheme", "authorization", "Bearer abc.def", "Bearer ***"},
		{"authorization without scheme", "Authorization", probeToken, "***"},
		{"api key", "X-Api-Key", "abcdef123456", "***"},
		{"token header", "X-Auth-Token", "zzz", "***"},
		{"ordinary header", "Content-Type", "application/json", "application/json"},
	}
	l := NewLogger(DefaultLoggerConfig())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := l.SanitizeValue(tc.key, tc.value); got != tc.want {
				t.Errorf("SanitizeValue(%q, %q) = %q, want %q", tc.key, tc.value, got, tc.want)
			}
		})
	}
}
