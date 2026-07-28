package share

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	Prefix  = "HILT"
	Version = 1

	KindAdversary   = "adversary"
	KindEnvironment = "environment"

	maxDecoded = 1 << 20
)

type Payload struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

func Encode(kind string, v any) (string, error) {
	if kind != KindAdversary && kind != KindEnvironment {
		return "", fmt.Errorf("unknown share kind %q", kind)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encoding %s: %w", kind, err)
	}
	body, err := json.Marshal(Payload{Kind: kind, Data: data})
	if err != nil {
		return "", fmt.Errorf("encoding payload: %w", err)
	}

	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return "", fmt.Errorf("compressing: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return "", fmt.Errorf("compressing: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("compressing: %w", err)
	}

	return fmt.Sprintf("%s%d:%s", Prefix, Version, base64.RawURLEncoding.EncodeToString(buf.Bytes())), nil
}

func Decode(code string) (Payload, error) {
	code = strings.TrimSpace(code)
	code = strings.Join(strings.Fields(code), "")
	if code == "" {
		return Payload{}, fmt.Errorf("paste a share code first")
	}

	prefix, encoded, ok := strings.Cut(code, ":")
	if !ok || !strings.HasPrefix(prefix, Prefix) {
		return Payload{}, fmt.Errorf("that doesn't look like a Hilt share code")
	}

	var version int
	if _, err := fmt.Sscanf(strings.TrimPrefix(prefix, Prefix), "%d", &version); err != nil {
		return Payload{}, fmt.Errorf("that doesn't look like a Hilt share code")
	}
	if version > Version {
		return Payload{}, fmt.Errorf("this code is from a newer version of Hilt (format %d, this build reads %d)", version, Version)
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Payload{}, fmt.Errorf("this code is damaged — it may have been cut short when copied")
	}

	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return Payload{}, fmt.Errorf("this code is damaged — it may have been cut short when copied")
	}
	defer r.Close()

	body, err := io.ReadAll(io.LimitReader(r, maxDecoded+1))
	if err != nil {
		return Payload{}, fmt.Errorf("this code is damaged — it may have been cut short when copied")
	}
	if len(body) > maxDecoded {
		return Payload{}, fmt.Errorf("this code expands to more than %d bytes and was refused", maxDecoded)
	}

	var payload Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		return Payload{}, fmt.Errorf("this code is damaged — it may have been cut short when copied")
	}
	if payload.Kind != KindAdversary && payload.Kind != KindEnvironment {
		return Payload{}, fmt.Errorf("unknown share kind %q", payload.Kind)
	}
	return payload, nil
}

func (p Payload) Into(v any) error {
	if err := json.Unmarshal(p.Data, v); err != nil {
		return fmt.Errorf("reading shared %s: %w", p.Kind, err)
	}
	return nil
}
