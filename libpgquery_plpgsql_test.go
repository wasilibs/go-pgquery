package pg_query_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	pg_query "github.com/wasilibs/go-pgquery"
)

//go:embed libpgquery_plpgsql_samples.sql
var plpgsqlSamples string

//go:embed libpgquery_plpgsql_samples.expected.json
var plpgsqlSamplesExpected string

func TestLibPgqueryPlsql(t *testing.T) {
	actual, err := pg_query.ParsePlPgSqlToJSON(plpgsqlSamples)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var actualParsed []interface{}
	if err := json.Unmarshal([]byte(actual), &actualParsed); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var expectedParsed []interface{}
	if err := json.Unmarshal([]byte(plpgsqlSamplesExpected), &expectedParsed); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actualParsed, expectedParsed) {
		t.Errorf("expected %q, got %q", expectedParsed, actualParsed)
	}
}
