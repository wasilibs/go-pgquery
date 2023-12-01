package pg_query_test

import (
	"fmt"
	"strings"
	"testing"

	pg_query "github.com/wasilibs/go-pgquery"
)

// https://github.com/pganalyze/libpg_query/blob/15-4.2.3/test/scan_tests.c
var libpgqueryScanTests = []string{
	"SELECT 1",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"1 = ICONST, NO_KEYWORD\n",
	"SELECT * FROM x WHERE z = 2",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"* = ASCII_42, NO_KEYWORD\n" +
		"FROM = FROM, RESERVED_KEYWORD\n" +
		"x = IDENT, NO_KEYWORD\n" +
		"WHERE = WHERE, RESERVED_KEYWORD\n" +
		"z = IDENT, NO_KEYWORD\n" +
		"= = ASCII_61, NO_KEYWORD\n" +
		"2 = ICONST, NO_KEYWORD\n",
	"SELECT 5.41414",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"5.41414 = FCONST, NO_KEYWORD\n",
	"SELECT $1",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"$1 = PARAM, NO_KEYWORD\n",
	"SELECT update AS left /* comment */ FROM between",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"update = UPDATE, UNRESERVED_KEYWORD\n" +
		"AS = AS, RESERVED_KEYWORD\n" +
		"left = LEFT, TYPE_FUNC_NAME_KEYWORD\n" +
		"/* comment */ = C_COMMENT, NO_KEYWORD\n" +
		"FROM = FROM, RESERVED_KEYWORD\n" +
		"between = BETWEEN, COL_NAME_KEYWORD\n",
	"SELECT 999999999999999999999::numeric/1000000000000000000000",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"999999999999999999999 = FCONST, NO_KEYWORD\n" +
		":: = TYPECAST, NO_KEYWORD\n" +
		"numeric = NUMERIC, COL_NAME_KEYWORD\n" +
		"/ = ASCII_47, NO_KEYWORD\n" +
		"1000000000000000000000 = FCONST, NO_KEYWORD\n",
	"SELECT 4790999999999999999999999999999999999999999999999999999999999999999999999999999999999999 * 9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"4790999999999999999999999999999999999999999999999999999999999999999999999999999999999999 = FCONST, NO_KEYWORD\n" +
		"* = ASCII_42, NO_KEYWORD\n" +
		"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999 = FCONST, NO_KEYWORD\n",
	"SELECT $foo$test$foo$, '123' FROM \"xy\", abc",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"$foo$test$foo$ = SCONST, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"'123' = SCONST, NO_KEYWORD\n" +
		"FROM = FROM, RESERVED_KEYWORD\n" +
		"\"xy\" = IDENT, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"abc = IDENT, NO_KEYWORD\n",
	"SELECT '/*not a comment*/', E'\\134', B'10', X'12', -Infinity, U&\"d!0061t!+000061\" UESCAPE '!'",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"'/*not a comment*/' = SCONST, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"E'\\134' = SCONST, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"B'10' = BCONST, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"X'12' = XCONST, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"- = ASCII_45, NO_KEYWORD\n" +
		"Infinity = IDENT, NO_KEYWORD\n" +
		", = ASCII_44, NO_KEYWORD\n" +
		"U = UIDENT, NO_KEYWORD\n" +
		"UESCAPE = UESCAPE, UNRESERVED_KEYWORD\n" +
		"'!' = SCONST, NO_KEYWORD\n",
	"SELECT 'a'/*;*/'b'; SELECT ';';",
	"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"'a' = SCONST, NO_KEYWORD\n" +
		"/*;*/ = C_COMMENT, NO_KEYWORD\n" +
		"'b' = SCONST, NO_KEYWORD\n" +
		"; = ASCII_59, NO_KEYWORD\n" +
		"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"';' = SCONST, NO_KEYWORD\n" +
		"; = ASCII_59, NO_KEYWORD\n",
	"CREATE RULE x AS ON SELECT TO tbl DO (SELECT 1; SELECT 2)",
	"CREATE = CREATE, RESERVED_KEYWORD\n" +
		"RULE = RULE, UNRESERVED_KEYWORD\n" +
		"x = IDENT, NO_KEYWORD\n" +
		"AS = AS, RESERVED_KEYWORD\n" +
		"ON = ON, RESERVED_KEYWORD\n" +
		"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"TO = TO, RESERVED_KEYWORD\n" +
		"tbl = IDENT, NO_KEYWORD\n" +
		"DO = DO, RESERVED_KEYWORD\n" +
		"( = ASCII_40, NO_KEYWORD\n" +
		"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"1 = ICONST, NO_KEYWORD\n" +
		"; = ASCII_59, NO_KEYWORD\n" +
		"SELECT = SELECT, RESERVED_KEYWORD\n" +
		"2 = ICONST, NO_KEYWORD\n" +
		") = ASCII_41, NO_KEYWORD\n",
}

// https://github.com/pganalyze/libpg_query/blob/15-4.2.3/test/scan.c
func TestLibPgqueryScan(t *testing.T) {
	tests := libpgqueryScanTests
	for i := 0; i < len(tests); i += 2 {
		input := tests[i]
		expected := tests[i+1]
		t.Run("", func(t *testing.T) {
			actual, err := pg_query.Scan(input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			var buffer strings.Builder
			for j := 0; j < len(actual.Tokens); j++ {
				scanToken := actual.Tokens[j]
				tokenKind := scanToken.Token
				keywordKind := scanToken.KeywordKind
				_, _ = fmt.Fprintf(&buffer, "%s = %s, %s\n", tests[i][scanToken.Start:scanToken.End], tokenKind.String(), keywordKind.String())
			}

			if expected != buffer.String() {
				t.Errorf("expected: %s, actual: %s", expected, buffer.String())
			}
		})
	}
}
