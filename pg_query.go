package pg_query

import (
	pganalyze "github.com/pganalyze/pg_query_go/v6"
	"github.com/wasilibs/go-pgquery/parser"
	"google.golang.org/protobuf/proto"
)

func Scan(input string) (result *pganalyze.ScanResult, err error) {
	protobufScan, err := parser.ScanToProtobuf(input)
	if err != nil {
		return
	}
	result = &pganalyze.ScanResult{}
	err = proto.Unmarshal(protobufScan, result)
	return
}

// ParseToJSON - Parses the given SQL statement into a parse tree (JSON format).
func ParseToJSON(input string) (result string, err error) {
	return parser.ParseToJSON(input) //nolint:wrapcheck // Simple proxy method, and match upstream
}

// Parse the given SQL statement into a parse tree (Go struct format).
func Parse(input string) (tree *pganalyze.ParseResult, err error) {
	protobufTree, err := parser.ParseToProtobuf(input)
	if err != nil {
		return
	}

	tree = &pganalyze.ParseResult{}
	err = proto.Unmarshal(protobufTree, tree)
	return
}

// Deparses a given Go parse tree into a SQL statement.
func Deparse(tree *pganalyze.ParseResult) (output string, err error) {
	protobufTree, err := proto.Marshal(tree)
	if err != nil {
		return
	}

	output, err = parser.DeparseFromProtobuf(protobufTree)
	return
}

// ParsePlPgSqlToJSON - Parses the given PL/pgSQL function statement into a parse tree (JSON format).
func ParsePlPgSqlToJSON(input string) (result string, err error) { //nolint:revive // Match upstream method name
	return parser.ParsePlPgSqlToJSON(input) //nolint:wrapcheck // Simple proxy method, and match upstream
}

// Normalize the passed SQL statement to replace constant values with ? characters.
func Normalize(input string) (result string, err error) {
	return parser.Normalize(input) //nolint:wrapcheck // Simple proxy method, and match upstream
}

// Fingerprint - Fingerprint the passed SQL statement to a hex string.
func Fingerprint(input string) (result string, err error) {
	return parser.FingerprintToHexStr(input) //nolint:wrapcheck // Simple proxy method, and match upstream
}

// FingerprintToUInt64 - Fingerprint the passed SQL statement to a uint64.
func FingerprintToUInt64(input string) (result uint64, err error) {
	return parser.FingerprintToUInt64(input) //nolint:wrapcheck // Simple proxy method, and match upstream
}

// HashXXH3_64 - Helper method to run XXH3 hash function (64-bit variant) on the given bytes, with the specified seed.
func HashXXH3_64(input []byte, seed uint64) (result uint64) {
	return parser.HashXXH3_64(input, seed)
}
