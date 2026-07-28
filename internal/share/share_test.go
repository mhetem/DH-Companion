package share

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mhetem/DH-Companion/internal/cards"
)

func rawCode(body []byte) (string, error) { return rawCodeVersion(Version, body) }

func rawCodeVersion(version int, body []byte) (string, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := w.Write(body); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%d:%s", Prefix, version, base64.RawURLEncoding.EncodeToString(buf.Bytes())), nil
}

func sampleAdversary() cards.Adversary {
	return cards.Adversary{
		Meta: cards.Meta{
			Kind:        cards.KindAdversary,
			Slug:        "gutter-wraith",
			Name:        "Gutter Wraith",
			Tier:        "2",
			Type:        "Standard",
			Description: "A drowned thing that remembers the tide.",
		},
		HordeNumber:    "1",
		Motives:        "Drag under, mourn",
		Experiences:    "Tidal Sense +2",
		Difficulty:     "15",
		ThresholdMinor: "7",
		ThresholdMajor: "14",
		Hp:             "6",
		Stress:         "4",
		StandardAttack: cards.Attack{
			Modifier:   "+2",
			Name:       "Grasp",
			Range:      "Melee",
			Damage:     "1d10+3",
			DamageType: "phy",
		},
		Features: []cards.Feature{
			{Title: "Undertow", Type: "Action", Description: "Pull a target to <strong>Close</strong> range."},
			{Title: "Drowned", Type: "Passive", Description: "Immune to drowning.\nMoves freely in water."},
		},
	}
}

func sampleEnvironment() cards.Environment {
	return cards.Environment{
		Meta: cards.Meta{
			Kind:        cards.KindEnvironment,
			Slug:        "abandoned-grove",
			Name:        "Abandoned Grove",
			Tier:        "1",
			Type:        "Exploration",
			Description: "A former druidic grove reclaimed by nature.",
		},
		Difficulty:           "11",
		Impulses:             "Draw in the curious, echo the past",
		PotentialAdversaries: []string{"Bear", "Dire Wolf", "Young Dryad"},
		Features: []cards.Feature{
			{Title: "Overgrown Battlefield", Type: "Passive", Description: "There has been a battle here."},
		},
	}
}

func TestEncodeDecodeAdversary(t *testing.T) {
	in := sampleAdversary()

	code, err := Encode(KindAdversary, in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	payload, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if payload.Kind != KindAdversary {
		t.Errorf("kind = %q, want %q", payload.Kind, KindAdversary)
	}

	var out cards.Adversary
	if err := payload.Into(&out); err != nil {
		t.Fatalf("Into: %v", err)
	}

	if out.Name != in.Name {
		t.Errorf("name = %q, want %q", out.Name, in.Name)
	}
	if out.Slug != in.Slug {
		t.Errorf("slug = %q, want %q", out.Slug, in.Slug)
	}
	if out.Difficulty != in.Difficulty || out.Hp != in.Hp || out.Stress != in.Stress {
		t.Errorf("stats = %q/%q/%q, want %q/%q/%q",
			out.Difficulty, out.Hp, out.Stress, in.Difficulty, in.Hp, in.Stress)
	}
	if out.StandardAttack != in.StandardAttack {
		t.Errorf("attack = %+v, want %+v", out.StandardAttack, in.StandardAttack)
	}
	if !reflect.DeepEqual(out.Features, in.Features) {
		t.Errorf("features = %+v, want %+v", out.Features, in.Features)
	}
}

func TestEncodeDecodeEnvironment(t *testing.T) {
	in := sampleEnvironment()

	code, err := Encode(KindEnvironment, in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	payload, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if payload.Kind != KindEnvironment {
		t.Errorf("kind = %q, want %q", payload.Kind, KindEnvironment)
	}

	var out cards.Environment
	if err := payload.Into(&out); err != nil {
		t.Fatalf("Into: %v", err)
	}
	if out.Name != in.Name || out.Impulses != in.Impulses {
		t.Errorf("got %q/%q, want %q/%q", out.Name, out.Impulses, in.Name, in.Impulses)
	}
	if len(out.PotentialAdversaries) != len(in.PotentialAdversaries) {
		t.Fatalf("potential adversaries = %d, want %d",
			len(out.PotentialAdversaries), len(in.PotentialAdversaries))
	}
	for i := range in.PotentialAdversaries {
		if out.PotentialAdversaries[i] != in.PotentialAdversaries[i] {
			t.Errorf("potential adversary %d = %q, want %q",
				i, out.PotentialAdversaries[i], in.PotentialAdversaries[i])
		}
	}
}

func TestEncodeRejectsUnknownKind(t *testing.T) {
	for _, kind := range []string{"", "character", "domain", "Adversary"} {
		if _, err := Encode(kind, sampleAdversary()); err == nil {
			t.Errorf("Encode(%q) = nil error, want rejection", kind)
		}
	}
}

func TestEncodedCodeShape(t *testing.T) {
	code, err := Encode(KindAdversary, sampleAdversary())
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if !strings.HasPrefix(code, "HILT1:") {
		t.Errorf("code = %q..., want HILT1: prefix", code[:min(12, len(code))])
	}
	if strings.ContainsAny(code, "+/=") {
		t.Error("code contains +, / or =, so it is not url-safe")
	}
	if strings.ContainsAny(code, " \t\r\n") {
		t.Error("code contains whitespace")
	}
}

func TestDecodeToleratesWhitespace(t *testing.T) {
	code, err := Encode(KindAdversary, sampleAdversary())
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	cut := len(code) / 2
	wrapped := []string{
		"  " + code + "\n",
		code[:cut] + "\n" + code[cut:],
		code[:10] + " " + code[10:cut] + "\r\n\t" + code[cut:],
		"\n\n" + code[:5] + "\n" + code[5:20] + "\n" + code[20:] + "\n",
	}

	for i, w := range wrapped {
		payload, err := Decode(w)
		if err != nil {
			t.Errorf("case %d: Decode: %v", i, err)
			continue
		}
		if payload.Kind != KindAdversary {
			t.Errorf("case %d: kind = %q, want %q", i, payload.Kind, KindAdversary)
		}
	}
}

func TestDecodeRejects(t *testing.T) {
	valid, err := Encode(KindAdversary, sampleAdversary())
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	tests := []struct {
		name string
		code string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t "},
		{"no separator", "HILT1"},
		{"plain text", "hello world"},
		{"wrong prefix", "NOPE1:abcd"},
		{"lowercase prefix", "hilt1:abcd"},
		{"no version", "HILT:abcd"},
		{"non-numeric version", "HILTx:abcd"},
		{"future version", "HILT9:abcd"},
		{"invalid base64", "HILT1:!!!!!!"},
		{"not zlib", "HILT1:aGVsbG8"},
		{"truncated", valid[:len(valid)-12]},
		{"body chopped", valid[:len(valid)/2]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decode(tt.code); err == nil {
				t.Errorf("Decode(%q) = nil error, want rejection", tt.code)
			}
		})
	}
}

func TestDecodeRejectsFutureVersion(t *testing.T) {
	body, err := json.Marshal(Payload{Kind: KindAdversary, Data: json.RawMessage(`{"name":"Gutter Wraith"}`)})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	code, err := rawCodeVersion(Version+1, body)
	if err != nil {
		t.Fatalf("rawCodeVersion: %v", err)
	}

	if _, err := Decode(code); err == nil {
		t.Fatal("Decode accepted a code from a newer format version")
	} else if !strings.Contains(err.Error(), "newer version of Hilt") {
		t.Errorf("error = %q, want the version gate rather than an incidental parse failure", err)
	}

	same, err := rawCodeVersion(Version, body)
	if err != nil {
		t.Fatalf("rawCodeVersion: %v", err)
	}
	if _, err := Decode(same); err != nil {
		t.Fatalf("the same payload at the current version must decode, got %v", err)
	}
}

func TestDecodeRejectsUnknownPayloadKind(t *testing.T) {
	code, err := Encode(KindAdversary, sampleAdversary())
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	payload, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	payload.Kind = "character"

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	forged, err := rawCode(body)
	if err != nil {
		t.Fatalf("rawCode: %v", err)
	}

	if _, err := Decode(forged); err == nil {
		t.Error("Decode accepted a payload with an unknown kind")
	}
}

func TestDecodeRefusesOversizedPayload(t *testing.T) {
	huge := sampleAdversary()
	huge.Description = strings.Repeat("a", 2*maxDecoded)

	code, err := Encode(KindAdversary, huge)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(code) > maxDecoded {
		t.Fatalf("code is %d bytes, so this does not exercise the inflate guard", len(code))
	}

	_, err = Decode(code)
	if err == nil {
		t.Fatal("Decode accepted a payload that inflates past the limit")
	}
	if !strings.Contains(err.Error(), "expands to more than") {
		t.Errorf("error = %q, want the size guard rather than an incidental parse failure", err)
	}
}

func TestIntoRejectsMismatchedData(t *testing.T) {
	payload := Payload{Kind: KindAdversary, Data: json.RawMessage(`"not an object"`)}

	var out cards.Adversary
	if err := payload.Into(&out); err == nil {
		t.Error("Into accepted data that does not fit the target")
	}
}

func TestRoundTripIsStable(t *testing.T) {
	in := sampleAdversary()

	first, err := Encode(KindAdversary, in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	payload, err := Decode(first)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var mid cards.Adversary
	if err := payload.Into(&mid); err != nil {
		t.Fatalf("Into: %v", err)
	}
	second, err := Encode(KindAdversary, mid)
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}

	if first != second {
		t.Error("a decoded card does not re-encode to the same code")
	}
}
