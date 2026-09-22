package goxz

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestArchivesAreReproducible(t *testing.T) {
	timestamp := time.Date(2000, time.January, 2, 3, 4, 6, 0, time.UTC)
	first := archiveFixture(t, "first", timestamp.Add(-time.Hour))
	second := archiveFixture(t, "second", timestamp.Add(time.Hour))

	tests := []struct {
		name    string
		archive func(string, io.Writer, time.Time) error
		check   func(*testing.T, []byte, time.Time)
	}{
		{"zip", archiveZip, checkZipArchive},
		{"tar.gz", archiveTarGz, checkTarGzArchive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var firstArchive, secondArchive bytes.Buffer
			if err := tt.archive(first, &firstArchive, timestamp); err != nil {
				t.Fatal(err)
			}
			if err := tt.archive(second, &secondArchive, timestamp); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(firstArchive.Bytes(), secondArchive.Bytes()) {
				t.Fatal("archives differ for identical content with different source mtimes")
			}
			tt.check(t, firstArchive.Bytes(), timestamp)
		})
	}
}

func archiveFixture(t *testing.T, parent string, mtime time.Time) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), parent, "package")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("readme\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "tool"), []byte("binary\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chtimes(path, mtime, mtime)
	}); err != nil {
		t.Fatal(err)
	}
	return dir
}

func checkZipArchive(t *testing.T, contents []byte, timestamp time.Time) {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		names = append(names, file.Name)
		if !file.Modified.Equal(timestamp) {
			t.Errorf("%s: modification time = %s, want %s", file.Name, file.Modified, timestamp)
		}
		if want := expectedArchiveMode(file.Name); file.Mode().Perm() != want {
			t.Errorf("%s: mode = %o, want %o", file.Name, file.Mode().Perm(), want)
		}
	}
	assertArchiveNames(t, names, []string{"package/", "package/README", "package/bin/", "package/bin/tool"})
}

func checkTarGzArchive(t *testing.T, contents []byte, timestamp time.Time) {
	t.Helper()
	gzipReader, err := gzip.NewReader(bytes.NewReader(contents))
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()
	if !gzipReader.ModTime.Equal(timestamp) {
		t.Errorf("gzip modification time = %s, want %s", gzipReader.ModTime, timestamp)
	}
	tarReader := tar.NewReader(gzipReader)
	var names []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, header.Name)
		if !header.ModTime.Equal(timestamp) {
			t.Errorf("%s: modification time = %s, want %s", header.Name, header.ModTime, timestamp)
		}
		if header.Uid != 0 || header.Gid != 0 || header.Uname != "" || header.Gname != "" {
			t.Errorf("%s: owner = %d:%d %q:%q, want 0:0 empty names",
				header.Name, header.Uid, header.Gid, header.Uname, header.Gname)
		}
		if want := int64(expectedArchiveMode(header.Name)); header.Mode != want {
			t.Errorf("%s: mode = %o, want %o", header.Name, header.Mode, want)
		}
	}
	assertArchiveNames(t, names, []string{"package", "package/README", "package/bin", "package/bin/tool"})
}

func assertArchiveNames(t *testing.T, names, want []string) {
	t.Helper()
	if !reflect.DeepEqual(names, want) {
		t.Errorf("entry order = %v, want %v", names, want)
	}
}

func expectedArchiveMode(name string) os.FileMode {
	if name == "package" || name == "package/" || name == "package/bin" || name == "package/bin/" || name == "package/bin/tool" {
		return 0755
	}
	return 0644
}

func TestArchiveTimestampUsesSourceDateEpoch(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "946782246")
	timestamp, err := archiveTimestamp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2000, time.January, 2, 3, 4, 6, 0, time.UTC)
	if !timestamp.Equal(want) {
		t.Errorf("timestamp = %s, want %s", timestamp, want)
	}
}

func TestArchiveTimestampFallsBackWhenUnavailable(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "")
	timestamp, err := archiveTimestamp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !timestamp.IsZero() {
		t.Errorf("timestamp = %s, want zero time", timestamp)
	}
}
