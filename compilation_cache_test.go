//go:build !pgquery_cgo && !tinygo

package pg_query_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	wasilibs_pg_query "github.com/wasilibs/go-pgquery"
)

// cacheChildDir names the cache directory of a child process
// TestSetCompilationCacheDir starts. The cache must be set before the first
// parse of a process, which other tests of this package may already have
// done, so each run that sets it is a new process.
const cacheChildDir = "GO_PGQUERY_TEST_CACHE_DIR"

func TestSetCompilationCacheDir(t *testing.T) {
	if dir := os.Getenv(cacheChildDir); dir != "" {
		if err := wasilibs_pg_query.SetCompilationCacheDir(dir); err != nil {
			fmt.Fprintln(os.Stderr, "set before parsing:", err)
			os.Exit(1)
		}
		if _, err := wasilibs_pg_query.Parse("SELECT 1"); err != nil {
			fmt.Fprintln(os.Stderr, "parse:", err)
			os.Exit(1)
		}
		if err := wasilibs_pg_query.SetCompilationCacheDir(dir); err == nil {
			fmt.Fprintln(os.Stderr, "set after parsing: no error")
			os.Exit(1)
		}
		os.Exit(0)
	}

	dir := t.TempDir()
	run := func() time.Duration {
		t.Helper()
		cmd := exec.Command(os.Args[0], "-test.run=^TestSetCompilationCacheDir$")
		cmd.Env = append(os.Environ(), cacheChildDir+"="+dir)
		start := time.Now()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("child: %v\n%s", err, out)
		}
		return time.Since(start)
	}

	first := run()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the first run left nothing in the cache directory")
	}
	second := run()
	t.Logf("first run %v, second run with the cache %v", first, second)
}

func TestSetCompilationCacheDirBadDir(t *testing.T) {
	file := t.TempDir() + "/file"
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := wasilibs_pg_query.SetCompilationCacheDir(file); err == nil {
		t.Fatal("a file as the cache directory: no error")
	}
}
