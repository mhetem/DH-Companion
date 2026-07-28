package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want [3]int
		ok   bool
	}{
		{"v1.2.3", [3]int{1, 2, 3}, true},
		{"1.2.3", [3]int{1, 2, 3}, true},
		{"v1.2", [3]int{1, 2, 0}, true},
		{"v1", [3]int{1, 0, 0}, true},
		{"  v1.2.3  ", [3]int{1, 2, 3}, true},
		{"v0.0.1", [3]int{0, 0, 1}, true},
		{"v10.20.30", [3]int{10, 20, 30}, true},
		{"v1.2.3-beta.1", [3]int{1, 2, 3}, true},
		{"v1.2.3+build7", [3]int{1, 2, 3}, true},
		{"", [3]int{}, false},
		{"v", [3]int{}, false},
		{"dev", [3]int{}, false},
		{"latest", [3]int{}, false},
		{"v1.2.3.4", [3]int{}, false},
		{"v1.a.3", [3]int{}, false},
		{"v-1.2.3", [3]int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := parse(tt.in)
			if ok != tt.ok {
				t.Fatalf("parse(%q) ok = %v, want %v", tt.in, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		known   bool
		newer   bool
	}{
		{"patch behind", "v1.2.3", "v1.2.4", true, true},
		{"minor behind", "v1.2.9", "v1.3.0", true, true},
		{"major behind", "v1.9.9", "v2.0.0", true, true},
		{"identical", "v1.2.3", "v1.2.3", true, false},
		{"equivalent shorthand", "v1.0.0", "v1", true, false},
		{"ahead", "v2.0.0", "v1.9.9", true, false},
		{"prerelease ignored", "v1.2.3", "v1.2.3-rc1", true, false},
		{"dev build", "dev", "v1.2.3", false, false},
		{"unparseable latest", "v1.2.3", "nightly", false, false},
		{"both unparseable", "dev", "nightly", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			known, newer := compare(tt.current, tt.latest)
			if known != tt.known || newer != tt.newer {
				t.Errorf("compare(%q, %q) = known %v newer %v, want known %v newer %v",
					tt.current, tt.latest, known, newer, tt.known, tt.newer)
			}
		})
	}
}

func feedServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCheckAtReportsNewerRelease(t *testing.T) {
	srv := feedServer(t, http.StatusOK,
		`{"tag_name":"v1.3.0","html_url":"https://example.invalid/r/v1.3.0","body":"notes here"}`)

	got, err := checkAt(context.Background(), srv.URL, "v1.2.0")
	if err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if got.Current != "v1.2.0" || got.Latest != "v1.3.0" {
		t.Errorf("versions = %q/%q, want v1.2.0/v1.3.0", got.Current, got.Latest)
	}
	if !got.Known || !got.Newer {
		t.Errorf("known = %v, newer = %v, want both true", got.Known, got.Newer)
	}
	if got.URL != "https://example.invalid/r/v1.3.0" {
		t.Errorf("url = %q", got.URL)
	}
	if got.Notes != "notes here" {
		t.Errorf("notes = %q", got.Notes)
	}
}

func TestCheckAtWhenUpToDate(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{"tag_name":"v1.2.0"}`)

	got, err := checkAt(context.Background(), srv.URL, "v1.2.0")
	if err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if !got.Known {
		t.Error("known = false, want true")
	}
	if got.Newer {
		t.Error("newer = true for an identical version")
	}
}

func TestCheckAtWhenAhead(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{"tag_name":"v1.0.0"}`)

	got, err := checkAt(context.Background(), srv.URL, "v1.4.0")
	if err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if !got.Known || got.Newer {
		t.Errorf("known = %v, newer = %v, want true/false", got.Known, got.Newer)
	}
}

func TestCheckAtDevBuildIsNotComparable(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{"tag_name":"v1.3.0"}`)

	got, err := checkAt(context.Background(), srv.URL, "dev")
	if err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if got.Known {
		t.Error("known = true for a dev build")
	}
	if got.Newer {
		t.Error("newer = true for a dev build, which would nag on every unreleased build")
	}
	if got.Latest != "v1.3.0" {
		t.Errorf("latest = %q, want v1.3.0", got.Latest)
	}
}

func TestCheckAtTrimsTagWhitespace(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{"tag_name":"  v1.3.0\n"}`)

	got, err := checkAt(context.Background(), srv.URL, "v1.2.0")
	if err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if got.Latest != "v1.3.0" {
		t.Errorf("latest = %q, want v1.3.0", got.Latest)
	}
	if !got.Newer {
		t.Error("newer = false, want true")
	}
}

func TestCheckAtErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"no releases", http.StatusNotFound, `{}`, "no releases published yet"},
		{"rate limited", http.StatusForbidden, `{}`, "rate-limited"},
		{"server error", http.StatusInternalServerError, `{}`, "GitHub answered"},
		{"unreadable body", http.StatusOK, `not json at all`, "unreadable release feed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := feedServer(t, tt.status, tt.body)

			got, err := checkAt(context.Background(), srv.URL, "v1.2.0")
			if err == nil {
				t.Fatalf("checkAt = nil error, want one containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to contain %q", err, tt.want)
			}
			if got.Current != "v1.2.0" {
				t.Errorf("current = %q, want it preserved on the error path", got.Current)
			}
		})
	}
}

func TestCheckAtUnreachable(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{}`)
	url := srv.URL
	srv.Close()

	if _, err := checkAt(context.Background(), url, "v1.2.0"); err == nil {
		t.Fatal("checkAt = nil error against a closed server")
	} else if !strings.Contains(err.Error(), "couldn't reach GitHub") {
		t.Errorf("error = %q, want the connection message", err)
	}
}

func TestCheckAtHonoursCancelledContext(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{"tag_name":"v1.3.0"}`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := checkAt(ctx, srv.URL, "v1.2.0"); err == nil {
		t.Fatal("checkAt = nil error with a cancelled context")
	}
}

func TestCheckAtCapsBodySize(t *testing.T) {
	body := `{"tag_name":"v1.3.0","body":"` + strings.Repeat("a", 2*maxBody) + `"}`
	srv := feedServer(t, http.StatusOK, body)

	if _, err := checkAt(context.Background(), srv.URL, "v1.2.0"); err == nil {
		t.Fatal("checkAt accepted a body past the read limit")
	}
}

func TestCheckAtSendsGitHubHeaders(t *testing.T) {
	var accept, agent, method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept, agent, method = r.Header.Get("Accept"), r.Header.Get("User-Agent"), r.Method
		fmt.Fprint(w, `{"tag_name":"v1.3.0"}`)
	}))
	defer srv.Close()

	if _, err := checkAt(context.Background(), srv.URL, "v1.2.0"); err != nil {
		t.Fatalf("checkAt: %v", err)
	}
	if method != http.MethodGet {
		t.Errorf("method = %q, want GET", method)
	}
	if accept != "application/vnd.github+json" {
		t.Errorf("Accept = %q", accept)
	}
	if agent != userAgent {
		t.Errorf("User-Agent = %q, want %q", agent, userAgent)
	}
}
