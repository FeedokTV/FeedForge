package normalize_test

import (
	"bytes"
	"context"
	"encoding/json"
	"feedforge/internal/normalize"
	"feedforge/internal/parse"
	"feedforge/internal/profile"
	"feedforge/internal/runtime"
	"flag"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden files")

func TestMapURLhaus(t *testing.T) {
	file, err := os.Open("../../testdata/samples/urlhaus_sample.csv")

	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	urlhausProfile, err := profile.LoadBuiltin("urlhaus")
	if err != nil {
		t.Fatal(err)
	}

	parser, err := parse.New("csv")
	if err != nil {
		t.Fatal(err)
	}

	rawRecords := parser.Parse(context.Background(), file)
	mappedRecords := normalize.Map(urlhausProfile, rawRecords, runtime.NewStats())

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	for record, err := range mappedRecords {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if err := enc.Encode(record); err != nil {
			t.Fatal(err)
		}
	}

	goldenPath := "../../testdata/golden/urlhaus_sample.ndjson"

	// Update flag - regenerates golden file
	if *update {
		err := os.WriteFile(goldenPath, buf.Bytes(), 0644)
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("output differs from golden file")
	}
}
