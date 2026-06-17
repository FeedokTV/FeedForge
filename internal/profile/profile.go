package profile

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var ValidFormats = []string{"csv", "jsonl", "list"}

var KnownCanonical = []string{
	"value",
	"type",
	"source",
	"first_seen",
	"last_seen",
	"tags",
	"confidence",
}

type Profile struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Format        string          `yaml:"format"` // "csv" | "jsonl" | "list"
	Columns       []ColumnMap     `yaml:"columns"`
	Defaults      ProfileDefaults `yaml:"defaults"`
	ConfidenceMap map[string]int  `yaml:"confidence_map"`
}
type ColumnMap struct {
	Name       string `yaml:"name"`        // source column name
	Canonical  string `yaml:"canonical"`   // target field
	TypeHint   string `yaml:"type_hint"`   // "auto" | "url" | "ip" | ""
	TimeFormat string `yaml:"time_format"` // Go ref-time format
}

type ProfileDefaults struct {
	Source     string   `yaml:"source"`
	Tags       []string `yaml:"tags"`
	Confidence int      `yaml:"confidence"`
}

func (p *Profile) validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}

	if !slices.Contains(ValidFormats, p.Format) {
		return fmt.Errorf("format %s is not valid", p.Format)
	}

	for i, col := range p.Columns {
		if col.Name == "" {
			return fmt.Errorf("column num %d: name is required", i)
		}

		if col.Canonical == "" {
			return fmt.Errorf("column %q: canonical is required", col.Canonical)
		}

		if !strings.HasPrefix(col.Canonical, "meta.") &&
			!slices.Contains(KnownCanonical, col.Canonical) {
			return fmt.Errorf("column %q: unknown canonical field %q", col.Name, col.Canonical)
		}
	}

	return nil
}

func loadFromBytes(profileBytes []byte) (*Profile, error) {
	var profile Profile

	err := yaml.Unmarshal(profileBytes, &profile)
	if err != nil {
		return nil, fmt.Errorf("cannot load yaml: %w", err)
	}

	if err := profile.validate(); err != nil {
		return nil, err
	}

	return &profile, nil
}

func Load(profilePath string) (*Profile, error) {
	var profile *Profile

	file, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	profile, err = loadFromBytes(fileBytes)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func LoadBuiltin(profileName string) (*Profile, error) {
	var profile *Profile
	var err error

	switch profileName {
	case "urlhaus":
		profile, err = loadFromBytes(BuiltinUrlhausProfile)
	case "openphish":
		profile, err = loadFromBytes(BuiltinOpenphishProfile)
	case "threatfox":
		profile, err = loadFromBytes(BuiltinThreatfoxProfile)
	case "generic-csv":
		profile, err = loadFromBytes(BuiltinGenericCSVProfile)
	case "generic-list":
		profile, err = loadFromBytes(BuiltinGenericListProfile)
	default:
		return nil, fmt.Errorf("builtin profile %q not found", profileName)
	}

	return profile, err
}
