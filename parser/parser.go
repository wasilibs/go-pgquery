package parser

type Error struct {
	Message   string // exception message
	Funcname  string // source function of exception (e.g. SearchSysCache)
	Filename  string // source of exception (e.g. parse.l)
	Lineno    int    // source of exception (e.g. 104)
	Cursorpos int    // char in query at which exception occurred
	Context   string // additional context (optional, can be NULL)
}

func (e *Error) Error() string {
	return e.Message
}

// ParseToJSON - Parses the given SQL statement into a parse tree (JSON format)
func ParseToJSON(input string) (result string, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryParse(inputC)
}

// ParseToProtobuf - Parses the given SQL statement into a parse tree (Protobuf format)
func ParseToProtobuf(input string) (result []byte, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryParseProtobuf(inputC)
}

// Scans the given SQL statement into a protobuf ScanResult
func ScanToProtobuf(input string) (result []byte, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryScanProtobuf(inputC)
}

// ParsePlPgSqlToJSON - Parses the given PL/pgSQL function statement into a parse tree (JSON format)
func ParsePlPgSqlToJSON(input string) (result string, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryParsePlPgSqlToJSON(inputC)
}

// Normalize the passed SQL statement to replace constant values with ? characters
func Normalize(input string) (result string, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryNormalize(inputC)
}

// FingerprintToUInt64 - Fingerprint the passed SQL statement using the C extension and returns result as uint64
func FingerprintToUInt64(input string) (result uint64, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryFingerprintToUint64(inputC)
}

// FingerprintToHexStr - Fingerprint the passed SQL statement using the C extension and returns result as hex string
func FingerprintToHexStr(input string) (result string, err error) {
	abi := newABI()
	defer abi.Close()

	inputC := abi.newCString(input)
	defer inputC.Close()

	return abi.pgQueryFingerprintToHexStr(inputC)
}
