// Package archive: write, read, copy, append, list primitives
// across all supported formats
/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */
package archive_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"testing"

	"github.com/NVIDIA/aistore/cmn/archive"
)

func TestReadOneEmptyArchiveName(t *testing.T) {
	tests := []struct {
		name string
		mime string
	}{
		{"tar", archive.ExtTar},
		{"zip", archive.ExtZip},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			switch test.mime {
			case archive.ExtTar:
				tw := tar.NewWriter(&buf)
				hdr := &tar.Header{Name: "", Mode: 0o600, Size: 3, Typeflag: tar.TypeReg}
				if err := tw.WriteHeader(hdr); err != nil {
					t.Fatal(err)
				}
				if _, err := tw.Write([]byte("abc")); err != nil {
					t.Fatal(err)
				}
				if err := tw.Close(); err != nil {
					t.Fatal(err)
				}
			case archive.ExtZip:
				zw := zip.NewWriter(&buf)
				w, err := zw.Create("")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := w.Write([]byte("abc")); err != nil {
					t.Fatal(err)
				}
				if err := zw.Close(); err != nil {
					t.Fatal(err)
				}
			}

			ar, err := archive.NewReader(test.mime, bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			if err != nil {
				t.Fatal(err)
			}
			r, err := ar.ReadOne("anything")
			if err != nil {
				t.Fatal(err)
			}
			if r != nil {
				_ = r.Close()
				t.Fatal("unexpected match for empty archived filename")
			}
		})
	}
}

func TestMimeContentType(t *testing.T) {
	tests := []struct {
		mime string
		ext  string
	}{
		// IANA and de-facto content types
		{"application/x-tar", archive.ExtTar},
		{"application/gzip", archive.ExtTgz},
		{"application/x-gzip", archive.ExtTgz},
		{"application/zip", archive.ExtZip},
		{"application/x-lz4", archive.ExtTarLz4},
		// media types are case-insensitive; parameters do not select the format
		{"Application/Gzip", archive.ExtTgz},
		{"Application/X-Lz4", archive.ExtTarLz4},
		{"application/zip; name=backup.gzip", archive.ExtZip},
		// "application/x-" + extension (see ais/test/archive_test.go)
		{"application/x-tgz", archive.ExtTgz},
		{"application/x-tar.gz", archive.ExtTarGz},
		{"application/x-zip", archive.ExtZip},
		{"application/x-tar.lz4", archive.ExtTarLz4},
		// extensions, with and without leading '.'
		{archive.ExtTar, archive.ExtTar},
		{archive.ExtTgz, archive.ExtTgz},
		{archive.ExtTarGz, archive.ExtTarGz},
		{archive.ExtZip, archive.ExtZip},
		{archive.ExtTarLz4, archive.ExtTarLz4},
		{"tgz", archive.ExtTgz},
		{"zip", archive.ExtZip},
	}
	for _, test := range tests {
		m, err := archive.Mime(test.mime, "")
		if err != nil {
			t.Errorf("Mime(%q): unexpected error: %v", test.mime, err)
			continue
		}
		if m != test.ext {
			t.Errorf("Mime(%q) = %q, expected %q", test.mime, m, test.ext)
		}
	}

	// not archive formats
	for _, mime := range []string{".gz", ".lz4", "lz4", "application/octet-stream"} {
		if m, err := archive.Mime(mime, ""); err == nil {
			t.Errorf("Mime(%q) = %q, expected error", mime, m)
		}
	}

	// round-trip
	for _, ext := range archive.FileExtensions {
		ct := archive.ContentTypeFromExt(ext)
		m, err := archive.Mime(ct, "")
		if err != nil {
			t.Errorf("Mime(%q) [from %q]: unexpected error: %v", ct, ext, err)
			continue
		}
		want := ext
		if ext == archive.ExtTarGz {
			want = archive.ExtTgz // same content type
		}
		if m != want {
			t.Errorf("Mime(%q) [from %q] = %q, expected %q", ct, ext, m, want)
		}
	}
}
