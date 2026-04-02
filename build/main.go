package main

import (
	"github.com/goyek/x/boot"
	"github.com/wasilibs/tools/tasks"
)

func main() {
	tasks.Define(tasks.Params{
		LibraryName: "pgquery",
		LibraryRepo: "pganalyze/pg_query_go",
	})
	boot.Main()
}
