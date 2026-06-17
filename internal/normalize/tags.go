package normalize

import (
	"slices"
	"strings"
)

func CanonicalTags(tags []string) []string {
	var canonicalTags []string

	for _, tag := range tags {

		if tag == "" {
			continue
		}

		tag = strings.ToLower(tag)
		tag = strings.ReplaceAll(tag, " ", "")

		tag = strings.ReplaceAll(tag, "-", "_")

		if slices.Contains(canonicalTags, tag) {
			continue
		}

		canonicalTags = append(canonicalTags, tag)
	}

	return canonicalTags
}
