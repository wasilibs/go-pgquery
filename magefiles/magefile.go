package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"

	"github.com/wasilibs/magefiles" // mage:import
)

func init() {
	magefiles.SetLibraryName("pgquery")
}

// GenerateProto generates Go stubs from the libpgquery proto file.
// We regenerate because even when Go packages are different, only
// a single instance of a type can be registered if with the same
// package and it should be allowed to import both this and pg_query_go
// for transition reasons.
func GenerateProto() error {
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}

	resp, err := http.Get("https://raw.githubusercontent.com/pganalyze/libpg_query/15-4.2.3/protobuf/pg_query.proto")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath.Join("build", "pg_query_wasilibs.proto"))
	if err != nil {
		return err
	}

	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		if s.Text() == "package pg_query;" {
			_, _ = f.WriteString("package pg_query_wasilibs;\n")
			continue
		}
		_, _ = f.WriteString(s.Text() + "\n")
	}

	if err := sh.RunV("go", "run", "github.com/curioswitch/protog/cmd@v0.3.0", "-I", "build", "--go_out=.", "--go_opt=Mpg_query_wasilibs.proto=/pg_query", "build/pg_query_wasilibs.proto"); err != nil {
		return err
	}

	return nil
}

func UpdateCParser() error {
	ref := "2c36edb70a84d3fa060f41080f599696ecebd8fd"
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
