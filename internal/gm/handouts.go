package gm

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

const MediaRoute = "/media/"

const defaultHandoutTitle = "Untitled"

var mediaTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".gif":  "image/gif",
	".mp4":  "video/mp4",
	".m4v":  "video/mp4",
	".webm": "video/webm",
}

var mediaFileName = regexp.MustCompile(`^[0-9a-f]{64}\.[a-z0-9]+$`)

func MediaContentType(file string) (string, bool) {
	if !mediaFileName.MatchString(file) {
		return "", false
	}
	typ, ok := mediaTypes[filepath.Ext(file)]
	return typ, ok
}

func MediaFilter() string {
	var patterns []string
	for _, ext := range slices.Sorted(maps.Keys(mediaTypes)) {
		patterns = append(patterns, "*"+ext, "*"+strings.ToUpper(ext))
	}
	return strings.Join(patterns, ";")
}

const (
	minViewZoom = 0.25
	maxViewZoom = 8
	maxViewPan  = 3
)

type HandoutView struct {
	Rotate int     `json:"rotate"`
	Zoom   float64 `json:"zoom"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

func finite(v, fallback float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fallback
	}
	return v
}

func NormalizeView(v HandoutView) HandoutView {
	rotate := ((v.Rotate%360)+360)%360 + 45
	zoom := finite(v.Zoom, 1)
	if zoom == 0 {
		zoom = 1
	}
	return HandoutView{
		Rotate: rotate / 90 * 90 % 360,
		Zoom:   min(max(zoom, minViewZoom), maxViewZoom),
		X:      min(max(finite(v.X, 0), -maxViewPan), maxViewPan),
		Y:      min(max(finite(v.Y, 0), -maxViewPan), maxViewPan),
	}
}

type Handout struct {
	ID           int64       `json:"id"`
	CampaignID   int64       `json:"campaignId"`
	Kind         string      `json:"kind"`
	Title        string      `json:"title"`
	Body         string      `json:"body"`
	MediaURL     string      `json:"mediaUrl"`
	MediaType    string      `json:"mediaType"`
	MediaName    string      `json:"mediaName"`
	MediaMissing bool        `json:"mediaMissing"`
	View         HandoutView `json:"view"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
}

type HandoutInput struct {
	ID         *int64 `json:"id"`
	CampaignID int64  `json:"campaignId"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Source     string `json:"source"`
	ClearMedia bool   `json:"clearMedia"`
}

func (in HandoutInput) validate() (HandoutInput, error) {
	if in.CampaignID <= 0 {
		return in, fmt.Errorf("a handout needs a campaign")
	}
	kind, err := validateHandoutKind(in.Kind)
	if err != nil {
		return in, err
	}
	in.Kind = kind
	in.Title = strings.TrimSpace(in.Title)
	in.Source = strings.TrimSpace(in.Source)
	if in.Source != "" {
		if _, ok := mediaTypes[strings.ToLower(filepath.Ext(in.Source))]; !ok {
			return in, fmt.Errorf("the player window shows PNG, JPEG, WebP and GIF images and MP4 or WebM video, not %s", filepath.Base(in.Source))
		}
	}
	return in, nil
}

func handoutTitle(mediaName string) string {
	if title := strings.TrimSpace(strings.TrimSuffix(mediaName, filepath.Ext(mediaName))); title != "" {
		return title
	}
	return defaultHandoutTitle
}

func (s *Service) handoutView(r db.Handout) Handout {
	h := Handout{
		ID:         r.ID,
		CampaignID: r.CampaignID,
		Kind:       r.Kind,
		Title:      r.Title,
		Body:       r.Body,
		MediaName:  r.MediaName,
		View: HandoutView{
			Rotate: int(r.ViewRotate),
			Zoom:   r.ViewZoom,
			X:      r.ViewX,
			Y:      r.ViewY,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Media != "" {
		typ, ok := MediaContentType(r.Media)
		h.MediaMissing = !ok || !s.hasMedia(r.Media)
		if ok {
			h.MediaURL = MediaRoute + r.Media
			h.MediaType, _, _ = strings.Cut(typ, "/")
		}
	}
	return h
}

func (s *Service) ListHandouts(campaignID int64) ([]Handout, error) {
	rows, err := s.q.ListHandoutsForCampaign(s.ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("listing handouts: %w", err)
	}
	out := make([]Handout, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.handoutView(r))
	}
	return out, nil
}

func (s *Service) GetHandout(id int64) (Handout, error) {
	row, err := s.q.GetHandout(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Handout{}, notFound("handout", fmt.Sprint(id))
	}
	if err != nil {
		return Handout{}, fmt.Errorf("loading handout: %w", err)
	}
	return s.handoutView(row), nil
}

func (s *Service) SaveHandout(in HandoutInput) (Handout, error) {
	in, err := in.validate()
	if err != nil {
		return Handout{}, err
	}

	var current db.Handout
	if in.ID != nil {
		current, err = s.q.GetHandout(s.ctx, *in.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return Handout{}, notFound("handout", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return Handout{}, fmt.Errorf("loading handout: %w", err)
		}
	}

	media, name := current.Media, current.MediaName
	if in.ClearMedia {
		media, name = "", ""
	}
	if in.Kind != "text" && media == "" && in.Source == "" {
		return Handout{}, fmt.Errorf("a %s needs an image or a video", in.Kind)
	}
	if in.Source != "" {
		media, name, err = s.importMedia(in.Source)
		if err != nil {
			return Handout{}, err
		}
	}
	if in.Title == "" {
		in.Title = handoutTitle(name)
	}

	row, err := s.writeHandout(in, media, name)
	if err != nil {
		if media != current.Media {
			s.releaseMedia(media)
		}
		return Handout{}, err
	}
	if current.Media != media {
		s.releaseMedia(current.Media)
	}
	return s.handoutView(row), nil
}

func (s *Service) writeHandout(in HandoutInput, media, name string) (db.Handout, error) {
	if in.ID == nil {
		row, err := s.q.CreateHandout(s.ctx, db.CreateHandoutParams{
			CampaignID: in.CampaignID,
			Kind:       in.Kind,
			Title:      in.Title,
			Body:       in.Body,
			Media:      media,
			MediaName:  name,
		})
		if err != nil {
			return db.Handout{}, fmt.Errorf("creating handout: %w", err)
		}
		return row, nil
	}
	row, err := s.q.UpdateHandout(s.ctx, db.UpdateHandoutParams{
		Kind:      in.Kind,
		Title:     in.Title,
		Body:      in.Body,
		Media:     media,
		MediaName: name,
		ID:        *in.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return db.Handout{}, notFound("handout", fmt.Sprint(*in.ID))
	}
	if err != nil {
		return db.Handout{}, fmt.Errorf("updating handout: %w", err)
	}
	return row, nil
}

func (s *Service) SaveHandoutView(id int64, view HandoutView) error {
	view = NormalizeView(view)
	if err := s.q.SetHandoutView(s.ctx, db.SetHandoutViewParams{
		ViewRotate: int64(view.Rotate),
		ViewZoom:   view.Zoom,
		ViewX:      view.X,
		ViewY:      view.Y,
		ID:         id,
	}); err != nil {
		return fmt.Errorf("saving the handout's framing: %w", err)
	}
	return nil
}

func (s *Service) DeleteHandout(id int64) error {
	row, err := s.q.GetHandout(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("loading handout: %w", err)
	}
	if err := s.q.DeleteHandout(s.ctx, id); err != nil {
		return fmt.Errorf("deleting handout: %w", err)
	}
	s.releaseMedia(row.Media)
	return nil
}

func (s *Service) importMedia(source string) (string, string, error) {
	name := filepath.Base(source)
	if s.media == "" {
		return "", "", fmt.Errorf("there's no media folder to copy %s into", name)
	}
	src, err := os.Open(source)
	if err != nil {
		return "", "", fmt.Errorf("opening %s: %w", name, err)
	}
	defer src.Close()

	if err := os.MkdirAll(s.media, 0o755); err != nil {
		return "", "", fmt.Errorf("creating the media folder: %w", err)
	}
	tmp, err := os.CreateTemp(s.media, ".import-*")
	if err != nil {
		return "", "", fmt.Errorf("copying %s: %w", name, err)
	}
	defer os.Remove(tmp.Name())

	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmp, hash), src)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", "", fmt.Errorf("copying %s: %w", name, err)
	}

	file := hex.EncodeToString(hash.Sum(nil)) + strings.ToLower(filepath.Ext(source))
	target := filepath.Join(s.media, file)
	if _, err := os.Stat(target); errors.Is(err, fs.ErrNotExist) {
		if err := os.Rename(tmp.Name(), target); err != nil {
			return "", "", fmt.Errorf("copying %s: %w", name, err)
		}
	} else if err != nil {
		return "", "", fmt.Errorf("copying %s: %w", name, err)
	}
	return file, name, nil
}

func (s *Service) hasMedia(file string) bool {
	if s.media == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(s.media, file))
	return err == nil
}

func (s *Service) releaseMedia(files ...string) {
	for _, file := range files {
		if s.media == "" || !mediaFileName.MatchString(file) {
			continue
		}
		if n, err := s.q.CountHandoutsUsingMedia(s.ctx, file); err != nil || n > 0 {
			continue
		}
		_ = os.Remove(filepath.Join(s.media, file))
	}
}
