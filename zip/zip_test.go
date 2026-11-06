package zip

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestZipPortableEntryNames(t *testing.T) {
	for _, insidePath := range []string{"src/", "nodejs/node_modules/"} {
		t.Run(insidePath, func(t *testing.T) {
			dir := t.TempDir()
			contents := map[string]string{
				"index.js":      "exports.handler = () => {};",
				"lib/helper.js": "exports.helper = true;",
			}
			var expectedFiles []string
			for name, content := range contents {
				filePath := filepath.Join(dir, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				expectedFiles = append(expectedFiles, filePath)
			}

			files, data, err := Zip(dir, insidePath)
			if err != nil {
				t.Fatal(err)
			}
			sort.Strings(files)
			sort.Strings(expectedFiles)
			if !reflect.DeepEqual(files, expectedFiles) {
				t.Errorf("filesystem paths = %q, want %q", files, expectedFiles)
			}

			reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			if len(reader.File) != len(contents) {
				t.Fatalf("archive has %d files, want %d", len(reader.File), len(contents))
			}
			var entryNames, expectedNames []string
			for _, entry := range reader.File {
				entryNames = append(entryNames, entry.Name)
			}
			for name := range contents {
				expectedNames = append(expectedNames, path.Join(insidePath, name))
			}
			sort.Strings(entryNames)
			sort.Strings(expectedNames)
			if !reflect.DeepEqual(entryNames, expectedNames) {
				t.Errorf("ZIP entry names = %q, want %q", entryNames, expectedNames)
			}
			for name, expectedContent := range contents {
				entryName := path.Join(insidePath, name)
				entry, err := reader.Open(entryName)
				if err != nil {
					t.Errorf("open portable archive entry %q: %s", entryName, err)
					continue
				}
				content, err := io.ReadAll(entry)
				entry.Close()
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != expectedContent {
					t.Errorf("content for %q = %q, want %q", entryName, content, expectedContent)
				}
			}
		})
	}
}
