package main

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"

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
