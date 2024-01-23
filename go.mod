module github.com/wasilibs/go-pgquery

go 1.19

require (
	github.com/google/go-cmp v0.5.5
	github.com/pganalyze/pg_query_go/v5 v5.1.0
	github.com/tetratelabs/wazero v1.3.0
	google.golang.org/protobuf v1.31.0
)

require golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect

replace github.com/tetratelabs/wazero => github.com/anuraaga/wazero v0.0.0-20240123011953-5c1baa8c3179
