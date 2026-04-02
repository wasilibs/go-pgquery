package pg_query_test

import (
	"testing"

	pganalyze "github.com/pganalyze/pg_query_go/v6"
	pg_query "github.com/wasilibs/go-pgquery"
	"github.com/wasilibs/go-pgquery/parser"
)

// Prevent compiler optimizations by assigning all results to global variables.
var (
	err        error
	resultStr  []byte
	resultTree *pganalyze.ParseResult
)

func benchmarkParse(b *testing.B, input string) {
	b.Helper()

	for range b.N {
		resultTree, err = pg_query.Parse(input)
		if err != nil {
			b.Errorf("Benchmark produced error %s\n\n", err)
		}
	}
}

func benchmarkParseParallel(b *testing.B, input string) {
	b.Helper()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = pg_query.Parse(input)
			if err != nil {
				b.Errorf("Benchmark produced error %s\n\n", err)
			}
		}
	})
}

func benchmarkRawParse(b *testing.B, input string) {
	b.Helper()

	for range b.N {
		resultStr, err = parser.ParseToProtobuf(input)
		if err != nil {
			b.Errorf("Benchmark produced error %s\n\n", err)
		}

		if len(resultStr) == 0 {
			b.Errorf("Benchmark produced empty result\n\n")
		}
	}
}

func benchmarkRawParseParallel(b *testing.B, input string) {
	b.Helper()

	b.RunParallel(func(pb *testing.PB) {
		var str []byte

		for pb.Next() {
			str, err = parser.ParseToProtobuf(input)
			if err != nil {
				b.Errorf("Benchmark produced error %s\n\n", err)
			}

			if len(str) == 0 {
				b.Errorf("Benchmark produced empty result\n\n")
			}
		}
	})
}

func benchmarkFingerprint(b *testing.B, input string) {
	b.Helper()

	var str string
	for range b.N {
		str, err = pg_query.Fingerprint(input)
		if err != nil {
			b.Errorf("Benchmark produced error %s\n\n", err)
		}
		if str == "" {
			b.Errorf("Benchmark produced empty result\n\n")
		}
	}
}

func benchmarkNormalize(b *testing.B, input string) {
	b.Helper()

	for range b.N {
		resultStr, err := pg_query.Normalize(input)
		if err != nil {
			b.Errorf("Benchmark produced error %s\n\n", err)
		}
		if resultStr == "" {
			b.Errorf("Benchmark produced empty result\n\n")
		}
	}
}

func BenchmarkParseSelect1(b *testing.B) {
	benchmarkParse(b, "SELECT 1")
}

func BenchmarkParseSelect2(b *testing.B) {
	benchmarkParse(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkParseCreateTable(b *testing.B) {
	benchmarkParse(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}

func BenchmarkParseSelect1Parallel(b *testing.B) {
	benchmarkParseParallel(b, "SELECT 1")
}

func BenchmarkParseSelect2Parallel(b *testing.B) {
	benchmarkParseParallel(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkParseCreateTableParallel(b *testing.B) {
	benchmarkParseParallel(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}

func BenchmarkRawParseSelect1(b *testing.B) {
	benchmarkRawParse(b, "SELECT 1")
}

func BenchmarkRawParseSelect2(b *testing.B) {
	benchmarkRawParse(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkRawParseCreateTable(b *testing.B) {
	benchmarkRawParse(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}

func BenchmarkRawParseSelect1Parallel(b *testing.B) {
	benchmarkRawParseParallel(b, "SELECT 1")
}

func BenchmarkRawParseSelect2Parallel(b *testing.B) {
	benchmarkRawParseParallel(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkRawParseCreateTableParallel(b *testing.B) {
	benchmarkRawParseParallel(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}

func BenchmarkFingerprintSelect1(b *testing.B) {
	benchmarkFingerprint(b, "SELECT 1")
}

func BenchmarkFingerprintSelect2(b *testing.B) {
	benchmarkFingerprint(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkFingerprintCreateTable(b *testing.B) {
	benchmarkFingerprint(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}

func BenchmarkNormalizeSelect1(b *testing.B) {
	benchmarkNormalize(b, "SELECT 1")
}

func BenchmarkNormalizeSelect2(b *testing.B) {
	benchmarkNormalize(b, "SELECT 1 FROM x WHERE y IN ('a', 'b', 'c')")
}

func BenchmarkNormalizeCreateTable(b *testing.B) {
	benchmarkNormalize(b, "CREATE TABLE types (a float(2), b float(49), c NUMERIC(2, 3), d character(4), e char(5), f varchar(6), g character varying(7))")
}
