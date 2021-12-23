package pkg

import (
	"strings"
	"testing"

	"github.com/anchore/packageurl-go"
	"github.com/go-test/deep"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestApkMetadata_pURL(t *testing.T) {
	tests := []struct {
		metadata ApkMetadata
		expected string
	}{
		{
			metadata: ApkMetadata{
				Package:      "p",
				Version:      "v",
				Architecture: "a",
			},
			expected: "pkg:alpine/p@v?arch=a",
		},
		// verify #351
		{
			metadata: ApkMetadata{
				Package:      "g++",
				Version:      "v84",
				Architecture: "am86",
			},
			expected: "pkg:alpine/g++@v84?arch=am86",
		},
		{
			metadata: ApkMetadata{
				Package:      "g plus plus",
				Version:      "v84",
				Architecture: "am86",
			},
			expected: "pkg:alpine/g%20plus%20plus@v84?arch=am86",
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			actual := test.metadata.PackageURL()
			if actual != test.expected {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(test.expected, actual, true)
				t.Errorf("diff: %s", dmp.DiffPrettyText(diffs))
			}
			// verify packageurl can parse
			purl, err := packageurl.FromString(actual)
			if err != nil {
				t.Errorf("cannot re-parse purl: %s", actual)
			}
			if purl.Name != test.metadata.Package {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(test.metadata.Package, purl.Name, true)
				t.Errorf("invalid purl name: %s", dmp.DiffPrettyText(diffs))
			}
			if purl.Version != test.metadata.Version {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(test.metadata.Version, purl.Version, true)
				t.Errorf("invalid purl version: %s", dmp.DiffPrettyText(diffs))
			}
			if purl.Qualifiers.Map()["arch"] != test.metadata.Architecture {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(test.metadata.Architecture, purl.Qualifiers.Map()["arch"], true)
				t.Errorf("invalid purl architecture: %s", dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestApkMetadata_FileOwner(t *testing.T) {
	tests := []struct {
		metadata ApkMetadata
		expected []string
	}{
		{
			metadata: ApkMetadata{
				Files: []ApkFileRecord{
					{Path: "/somewhere"},
					{Path: "/else"},
				},
			},
			expected: []string{
				"/else",
				"/somewhere",
			},
		},
		{
			metadata: ApkMetadata{
				Files: []ApkFileRecord{
					{Path: "/somewhere"},
					{Path: ""},
				},
			},
			expected: []string{
				"/somewhere",
			},
		},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.expected, ","), func(t *testing.T) {
			actual := test.metadata.OwnedFiles()
			for _, d := range deep.Equal(test.expected, actual) {
				t.Errorf("diff: %+v", d)
			}
		})
	}
}
