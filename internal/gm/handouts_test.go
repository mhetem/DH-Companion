package gm

import (
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeSource(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
	return path
}

func mediaFiles(t *testing.T, s *Service) []string {
	t.Helper()
	entries, err := os.ReadDir(s.media)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("reading media folder: %v", err)
	}
	out := []string{}
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func saveHandout(t *testing.T, s *Service, in HandoutInput) Handout {
	t.Helper()
	h, err := s.SaveHandout(in)
	if err != nil {
		t.Fatalf("SaveHandout(%+v): %v", in, err)
	}
	return h
}

func TestHandoutCopiesMediaAndTitlesFromFile(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	h := saveHandout(t, s, HandoutInput{
		CampaignID: c.ID,
		Kind:       "Map",
		Body:       "The secret door is on the north wall.",
		Source:     writeSource(t, "Ashfall Docks.PNG", "png bytes"),
	})

	if h.Kind != "map" || h.Title != "Ashfall Docks" || h.MediaName != "Ashfall Docks.PNG" {
		t.Fatalf("unexpected handout %+v", h)
	}
	if h.MediaType != "image" || h.MediaMissing || !strings.HasPrefix(h.MediaURL, MediaRoute) {
		t.Fatalf("unexpected media %+v", h)
	}
	file := strings.TrimPrefix(h.MediaURL, MediaRoute)
	if !strings.HasSuffix(file, ".png") {
		t.Fatalf("stored name %q should keep a lower-case extension", file)
	}
	if got := mediaFiles(t, s); !slices.Equal(got, []string{file}) {
		t.Fatalf("media folder holds %v, want just %s", got, file)
	}
	data, err := os.ReadFile(filepath.Join(s.media, file))
	if err != nil || string(data) != "png bytes" {
		t.Fatalf("copied file reads %q, %v", data, err)
	}
}

func TestHandoutNeedsMediaUnlessText(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	if _, err := s.SaveHandout(HandoutInput{CampaignID: c.ID, Kind: "scene", Title: "Nightfall"}); err == nil {
		t.Fatal("a scene without an image or video should be rejected")
	}
	if _, err := s.SaveHandout(HandoutInput{CampaignID: c.ID, Kind: "poster", Title: "Wanted"}); err == nil {
		t.Fatal("an unknown kind should be rejected")
	}
	if _, err := s.SaveHandout(HandoutInput{CampaignID: c.ID, Kind: "map", Source: writeSource(t, "map.svg", "<svg/>")}); err == nil {
		t.Fatal("an unsupported file type should be rejected")
	}
	if got := mediaFiles(t, s); len(got) != 0 {
		t.Fatalf("rejected saves left %v in the media folder", got)
	}

	text := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "text", Body: "The bridge is out."})
	if text.Title != defaultHandoutTitle || text.MediaURL != "" || text.MediaMissing {
		t.Fatalf("unexpected text handout %+v", text)
	}
}

func TestHandoutUpdateKeepsOrReplacesMedia(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	h := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "scene", Source: writeSource(t, "tavern.jpg", "still")})
	first := strings.TrimPrefix(h.MediaURL, MediaRoute)

	renamed := saveHandout(t, s, HandoutInput{ID: &h.ID, CampaignID: c.ID, Kind: "scene", Title: "The Ferryman's Rest"})
	if renamed.MediaURL != h.MediaURL || renamed.Title != "The Ferryman's Rest" {
		t.Fatalf("an update without a source should keep the media: %+v", renamed)
	}

	moving := saveHandout(t, s, HandoutInput{ID: &h.ID, CampaignID: c.ID, Kind: "scene", Title: renamed.Title, Source: writeSource(t, "tavern.webm", "moving")})
	if moving.MediaType != "video" || moving.MediaName != "tavern.webm" || moving.Title != "The Ferryman's Rest" {
		t.Fatalf("unexpected replacement %+v", moving)
	}
	second := strings.TrimPrefix(moving.MediaURL, MediaRoute)
	if got := mediaFiles(t, s); !slices.Equal(got, []string{second}) {
		t.Fatalf("media folder holds %v, want only the replacement %s (not %s)", got, second, first)
	}
}

func TestTextHandoutCanDropItsBackdrop(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	h := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "text", Title: "A letter", Source: writeSource(t, "parchment.webp", "paper")})
	if h.MediaType != "image" {
		t.Fatalf("text should accept a backdrop: %+v", h)
	}
	cleared := saveHandout(t, s, HandoutInput{ID: &h.ID, CampaignID: c.ID, Kind: "text", Title: "A letter", ClearMedia: true})
	if cleared.MediaURL != "" || cleared.MediaName != "" {
		t.Fatalf("backdrop should be gone: %+v", cleared)
	}
	if got := mediaFiles(t, s); len(got) != 0 {
		t.Fatalf("dropped backdrop left %v behind", got)
	}

	if _, err := s.SaveHandout(HandoutInput{ID: &h.ID, CampaignID: c.ID, Kind: "map", ClearMedia: true}); err == nil {
		t.Fatal("a map can't drop its image")
	}
}

func TestSharedMediaOutlivesOneHandout(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	a := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "map", Source: writeSource(t, "region.png", "same")})
	b := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "scene", Source: writeSource(t, "region.png", "same")})
	if a.MediaURL != b.MediaURL {
		t.Fatalf("identical files should share one copy: %s vs %s", a.MediaURL, b.MediaURL)
	}
	if got := mediaFiles(t, s); len(got) != 1 {
		t.Fatalf("media folder holds %v, want one shared file", got)
	}

	if err := s.DeleteHandout(a.ID); err != nil {
		t.Fatalf("DeleteHandout: %v", err)
	}
	if got := mediaFiles(t, s); len(got) != 1 {
		t.Fatalf("the file should stay while another handout uses it, have %v", got)
	}
	if err := s.DeleteHandout(b.ID); err != nil {
		t.Fatalf("DeleteHandout: %v", err)
	}
	if got := mediaFiles(t, s); len(got) != 0 {
		t.Fatalf("the last delete should remove the file, have %v", got)
	}
	if err := s.DeleteHandout(b.ID); err != nil {
		t.Fatalf("deleting twice should be a no-op, got %v", err)
	}
}

func TestDeletingCampaignReleasesItsMedia(t *testing.T) {
	s := newTestService(t)
	gone := newTestCampaign(t, s, "Age of Umbra")
	kept := newTestCampaign(t, s, "The Long Night")

	saveHandout(t, s, HandoutInput{CampaignID: gone.ID, Kind: "map", Source: writeSource(t, "only-here.png", "a")})
	saveHandout(t, s, HandoutInput{CampaignID: gone.ID, Kind: "scene", Source: writeSource(t, "both.png", "b")})
	shared := saveHandout(t, s, HandoutInput{CampaignID: kept.ID, Kind: "scene", Source: writeSource(t, "both.png", "b")})

	if err := s.DeleteCampaign(gone.ID); err != nil {
		t.Fatalf("DeleteCampaign: %v", err)
	}
	if got := mediaFiles(t, s); !slices.Equal(got, []string{strings.TrimPrefix(shared.MediaURL, MediaRoute)}) {
		t.Fatalf("media folder holds %v, want only the file the other campaign still uses", got)
	}
	left, err := s.ListHandouts(gone.ID)
	if err != nil || len(left) != 0 {
		t.Fatalf("handouts should go with their campaign, have %v, %v", left, err)
	}
}

func TestHandoutReportsMissingMedia(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	h := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "map", Source: writeSource(t, "lost.png", "x")})
	if err := os.Remove(filepath.Join(s.media, strings.TrimPrefix(h.MediaURL, MediaRoute))); err != nil {
		t.Fatalf("removing media: %v", err)
	}
	got, err := s.GetHandout(h.ID)
	if err != nil {
		t.Fatalf("GetHandout: %v", err)
	}
	if !got.MediaMissing {
		t.Fatalf("a handout whose file is gone should say so: %+v", got)
	}
}

func TestNormalizeView(t *testing.T) {
	for _, tc := range []struct {
		in, want HandoutView
	}{
		{HandoutView{}, HandoutView{Zoom: 1}},
		{HandoutView{Rotate: 90, Zoom: 2, X: 0.5, Y: -0.25}, HandoutView{Rotate: 90, Zoom: 2, X: 0.5, Y: -0.25}},
		{HandoutView{Rotate: -90, Zoom: 1}, HandoutView{Rotate: 270, Zoom: 1}},
		{HandoutView{Rotate: 450, Zoom: 1}, HandoutView{Rotate: 90, Zoom: 1}},
		{HandoutView{Rotate: 314, Zoom: 1}, HandoutView{Rotate: 270, Zoom: 1}},
		{HandoutView{Rotate: 316, Zoom: 1}, HandoutView{Rotate: 0, Zoom: 1}},
		{HandoutView{Zoom: 100, X: 9, Y: -9}, HandoutView{Zoom: maxViewZoom, X: maxViewPan, Y: -maxViewPan}},
		{HandoutView{Zoom: 0.01}, HandoutView{Zoom: minViewZoom}},
		{HandoutView{Zoom: math.NaN(), X: math.Inf(1)}, HandoutView{Zoom: 1}},
	} {
		if got := NormalizeView(tc.in); got != tc.want {
			t.Errorf("NormalizeView(%+v) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestHandoutViewRoundTrips(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	h := saveHandout(t, s, HandoutInput{CampaignID: c.ID, Kind: "map", Source: writeSource(t, "region.png", "x")})
	if h.View != (HandoutView{Zoom: 1}) {
		t.Fatalf("a new handout should start unframed, got %+v", h.View)
	}

	if err := s.SaveHandoutView(h.ID, HandoutView{Rotate: 90, Zoom: 1.5, X: 0.2, Y: -0.1}); err != nil {
		t.Fatalf("SaveHandoutView: %v", err)
	}
	renamed := saveHandout(t, s, HandoutInput{ID: &h.ID, CampaignID: c.ID, Kind: "map", Title: "The Drowned South"})
	want := HandoutView{Rotate: 90, Zoom: 1.5, X: 0.2, Y: -0.1}
	if renamed.View != want {
		t.Fatalf("editing the handout should keep its framing: got %+v, want %+v", renamed.View, want)
	}
}

func TestMediaContentTypeOnlyAcceptsStoredNames(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	for _, tc := range []struct {
		name string
		want bool
	}{
		{hash + ".png", true},
		{hash + ".webm", true},
		{hash + ".svg", false},
		{hash + ".PNG", false},
		{"../" + hash + ".png", false},
		{"tavern.png", false},
		{hash[:63] + ".png", false},
		{hash + ".png/../x.png", false},
		{strings.ToUpper(hash) + ".png", false},
	} {
		if _, ok := MediaContentType(tc.name); ok != tc.want {
			t.Errorf("MediaContentType(%q) = %v, want %v", tc.name, ok, tc.want)
		}
	}
}
