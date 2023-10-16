module github.com/wasilibs/go-pgquery

go 1.19

require (
	github.com/golang/protobuf v1.5.3
	github.com/magefile/mage v1.15.1-0.20230912152418-9f54e0f83e2a
	google.golang.org/protobuf v1.31.0
)

require github.com/tetratelabs/wazero v1.5.0 // indirect

replace github.com/tetratelabs/wazero v1.5.0 => ../wazero
