package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	feedURL   = "https://api.github.com/repos/mhetem/DH-Companion/releases/latest"
	userAgent = "Hilt"
	timeout   = 10 * time.Second
	maxBody   = 1 << 20
)

type Release struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	URL     string `json:"url"`
	Notes   string `json:"notes"`
	Newer   bool   `json:"newer"`
	Known   bool   `json:"known"`
}

type feed struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func Check(ctx context.Context, current string) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("building the update request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("couldn't reach GitHub — check your connection")
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Release{Current: current}, fmt.Errorf("no releases published yet")
	case resp.StatusCode == http.StatusForbidden:
		return Release{Current: current}, fmt.Errorf("GitHub rate-limited this check — try again later")
	case resp.StatusCode != http.StatusOK:
		return Release{Current: current}, fmt.Errorf("GitHub answered %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return Release{Current: current}, fmt.Errorf("reading the release feed: %w", err)
	}

	var latest feed
	if err := json.Unmarshal(body, &latest); err != nil {
		return Release{Current: current}, fmt.Errorf("unreadable release feed: %w", err)
	}

	out := Release{
		Current: current,
		Latest:  strings.TrimSpace(latest.TagName),
		URL:     latest.HTMLURL,
		Notes:   latest.Body,
	}
	out.Known, out.Newer = compare(current, out.Latest)
	return out, nil
}

func compare(current, latest string) (known, newer bool) {
	c, okC := parse(current)
	l, okL := parse(latest)
	if !okC || !okL {
		return false, false
	}
	for i := range c {
		if c[i] != l[i] {
			return true, l[i] > c[i]
		}
	}
	return true, false
}

func parse(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return out, false
	}
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
