package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wasilibs/magefiles" // mage:import
)

func init() {
	magefiles.SetLibraryName("pgquery")
}

func UpdateCParser() error {
	ref := "v6.1.0"
	uri := fmt.Sprintf("https://github.com/pganalyze/pg_query_go/archive/%s.zip", ref)
	res, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	srcZip, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(srcZip), int64(len(srcZip)))
	if err != nil {
		return err
	}

	if err := filepath.WalkDir(filepath.Join("internal", "cparser"), func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".go" {
			return nil
		}

		if err := os.Remove(path); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		pathParts := strings.Split(f.Name, string(filepath.Separator))
		if len(pathParts) < 2 {
			continue
		}
		if pathParts[1] != "parser" {
			continue
		}
		if filepath.Ext(f.Name) == ".go" {
			continue
		}
		outPath := filepath.Join(append([]string{"internal", "cparser"}, pathParts[2:]...)...)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		if _, err := io.Copy(out, rc); err != nil {
			return err
		}
	}

	return nil
}
