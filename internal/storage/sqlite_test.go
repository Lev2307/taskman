package storage

import (
	"slices"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tags := []string{"WOrk", "work", "spoRt", "urgent", "sport", "   Sport"}
	expectedTags := []string{"sport", "urgent", "work"}

	normTags := NormalizeTags(tags)
	if !slices.Equal(expectedTags, normTags) {
		t.Errorf("NormalizeTags() = %q, want %q", normTags, expectedTags)
	}
}
