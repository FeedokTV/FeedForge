package sink_test

import (
	"bytes"
	"feedforge/internal/helpers"
	"feedforge/internal/normalize"
	"feedforge/internal/sink"
	"feedforge/internal/testhelpers"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSinkNdjson(t *testing.T) {
	now := time.Date(2026, 6, 4, 23, 27, 26, 0, time.UTC)

	tests := []struct {
		Name    string
		Input   []normalize.Record
		Want    string
		WantErr bool
	}{
		{
			Name: "single record",
			Input: []normalize.Record{
				{
					ID:         "1",
					Type:       normalize.TypeURL,
					Source:     "test-source",
					Value:      "http://example.com",
					FirstSeen:  now,
					LastSeen:   &now,
					Tags:       []string{"tag1", "tag2"},
					Confidence: helpers.IntToPointer(90),
					Meta:       map[string]string{"key": "value"},
				},
			},
			Want: `{"id":"1","type":"url","source":"test-source","value":"http://example.com","first_seen":"2026-06-04T23:27:26Z","last_seen":"2026-06-04T23:27:26Z","tags":["tag1","tag2"],"confidence":90,"meta":{"key":"value"}}` + "\n",
		},
		{
			Name: "omitempty fields absent when zero",
			Input: []normalize.Record{
				{
					ID:        "2",
					Type:      normalize.TypeDomain,
					Source:    "test-source",
					Value:     "example.com",
					FirstSeen: now,
					// LastSeen, Tags, Confidence, Meta should be omitted
				},
			},
			Want: `{"id":"2","type":"domain","source":"test-source","value":"example.com","first_seen":"2026-06-04T23:27:26Z"}` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var buf bytes.Buffer

			err := sink.SaveNDJSON(&buf, testhelpers.RecordsToSeq(test.Input, nil))
			if test.WantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !test.WantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(test.Want, buf.String()); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
