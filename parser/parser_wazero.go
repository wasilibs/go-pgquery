//go:build !tinygo && !pgquery_cgo

package parser

import (
	"bytes"
	"container/list"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wasilibs/go-pgquery/internal/wasm"
	"github.com/wasilibs/wazero-helpers/allocator"
)

var (
	errFailedWrite = errors.New("failed to write to wasm memory")
	errFailedRead  = errors.New("failed to read from wasm memory")
)

var (
	wasmRT       wazero.Runtime
	wasmCompiled wazero.CompiledModule
	rootMod      api.Module

	wasmInitOnce sync.Once
)

func initWASM() {
	ctx := context.Background()
	ctx = experimental.WithMemoryAllocator(ctx, allocator.NewNonMoving())

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCoreFeatures(api.CoreFeaturesV2|experimental.CoreFeaturesThreads|experimental.CoreFeaturesExceptionHandling))

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	code, err := rt.CompileModule(ctx, wasm.LibPGQuery)
	if err != nil {
		panic(err)
	}

	if _, err := rt.InstantiateWithConfig(ctx, wasm.Memory, wazero.NewModuleConfig().WithName("env")); err != nil {
		panic(err)
	}

	cfg := wazero.NewModuleConfig().WithSysNanotime().WithStdout(os.Stdout).WithStderr(os.Stderr).WithName("").WithStartFunctions("_initialize")
	root, err := rt.InstantiateModule(ctx, code, cfg)
	if err != nil {
		panic(err)
	}

	wasmRT = rt
	wasmCompiled = code
	rootMod = root
}

// runWithABI borrows an ABI from the pool, runs fn, then either returns the ABI
// to the pool or discards it. SQL-level errors (*Error) leave the instance
// reusable; WASM/runtime failures and panics never recycle a possibly-corrupt
// instance (recycling after panic used to poison all later parses in-process).
func runWithABI(fn func(*abi) error) (err error) {
	abi := getABI()
	recycle := false
	defer func() {
		if r := recover(); r != nil {
			abi.discard()
			panic(r)
		}
		if recycle {
			abi.Close()
		} else {
			abi.discard()
		}
	}()
	err = fn(abi)
	recycle = shouldRecycleABI(err)
	return err
}

func shouldRecycleABI(err error) bool {
	if err == nil {
		return true
	}
	var pe *Error
	return errors.As(err, &pe)
}

// ParseToJSON - Parses the given SQL statement into a parse tree (JSON format).
func ParseToJSON(input string) (result string, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryParse(inputC)
		return err
	})
	return result, err
}

// ParseToProtobuf - Parses the given SQL statement into a parse tree (Protobuf format).
func ParseToProtobuf(input string) (result []byte, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryParseProtobuf(inputC)
		return err
	})
	return result, err
}

// DeparseFromProtobuf - Deparses the given Protobuf format parse tree into a SQL statement.
func DeparseFromProtobuf(input []byte) (result string, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCStringFromBytes(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryDeParseFromProtobuf(inputC)
		return err
	})
	return result, err
}

// Scans the given SQL statement into a protobuf ScanResult.
func ScanToProtobuf(input string) (result []byte, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryScanProtobuf(inputC)
		return err
	})
	return result, err
}

// ParsePlPgSqlToJSON - Parses the given PL/pgSQL function statement into a parse tree (JSON format).
func ParsePlPgSqlToJSON(input string) (result string, err error) { //nolint:revive // Match upstream method name
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryParsePlPgSqlToJSON(inputC)
		return err
	})
	return result, err
}

// Normalize the passed SQL statement to replace constant values with ? characters.
func Normalize(input string) (result string, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryNormalize(inputC)
		return err
	})
	return result, err
}

// FingerprintToUInt64 - Fingerprint the passed SQL statement using the C extension and returns result as uint64.
func FingerprintToUInt64(input string) (result uint64, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryFingerprintToUint64(inputC)
		return err
	})
	return result, err
}

// FingerprintToHexStr - Fingerprint the passed SQL statement using the C extension and returns result as hex string.
func FingerprintToHexStr(input string) (result string, err error) {
	err = runWithABI(func(abi *abi) error {
		inputC, err := abi.newCString(input)
		if err != nil {
			return err
		}
		defer inputC.Close()
		result, err = abi.pgQueryFingerprintToHexStr(inputC)
		return err
	})
	return result, err
}

// HashXXH3_64 - Helper method to run XXH3 hash function (64-bit variant) on the given bytes, with the specified seed.
func HashXXH3_64(input []byte, seed uint64) (result uint64) {
	abi := getABI()
	recycle := false
	defer func() {
		if r := recover(); r != nil {
			abi.discard()
			panic(r)
		}
		if recycle {
			abi.Close()
		} else {
			abi.discard()
		}
	}()

	inputC, err := abi.newCStringFromBytes(input)
	if err != nil {
		panic(err)
	}
	defer inputC.Close()

	result, err = abi.pgQueryHashXXH364(inputC, seed)
	if err != nil {
		panic(err)
	}
	recycle = true
	return result
}

var (
	abiPool   = list.New()
	abiPoolMu sync.Mutex

	// rootMu serializes execution of the root module, which shares one stack
	// and thread-local storage among its invocations.
	rootMu sync.Mutex
)

func newABI() *abi {
	ctx := context.Background()

	tlsSize := rootMod.ExportedGlobal("__tls_size").Get()
	tlsAlign := rootMod.ExportedGlobal("__tls_align").Get()
	// The stack is placed at the bottom of memory, so the root module's
	// initial stack pointer is also the size of the stack.
	stackSize := rootMod.ExportedGlobal("__stack_pointer").Get()
	size := tlsSize + tlsAlign + stackSize

	rootMu.Lock()
	mallocRes, err := rootMod.ExportedFunction("malloc").Call(ctx, size)
	rootMu.Unlock()
	if err != nil {
		panic(err)
	}
	ptr := uint32(mallocRes[0])

	cfg := wazero.NewModuleConfig().WithSysNanotime().WithStdout(os.Stdout).WithStderr(os.Stderr).WithName("").
		// Start functions already executed with the root module.
		WithStartFunctions()
	mod, err := wasmRT.InstantiateModule(ctx, wasmCompiled, cfg)
	if err != nil {
		panic(err)
	}

	tlsBase := (uint64(ptr) + tlsAlign - 1) &^ (tlsAlign - 1)
	if _, err := mod.ExportedFunction("__wasm_init_tls").Call(ctx, tlsBase); err != nil {
		_ = mod.Close(ctx)
		panic(err)
	}

	stackTop := (uint64(ptr) + size) &^ 15
	mod.ExportedGlobal("__stack_pointer").(api.MutableGlobal).Set(stackTop) //nolint:forcetypeassert // __stack_pointer is a mutable global in this module.

	res := &abi{
		fPgQueryInit:                    newLazyFunction(mod, "pg_query_init"),
		fPgQueryParse:                   newLazyFunction(mod, "pg_query_parse"),
		fPgQueryFreeParseResult:         newLazyFunction(mod, "pg_query_free_parse_result"),
		fPgQueryParseProtobuf:           newLazyFunction(mod, "pg_query_parse_protobuf"),
		fPgQueryFreeProtobufParseResult: newLazyFunction(mod, "pg_query_free_protobuf_parse_result"),
		fPgQueryParsePlpgsql:            newLazyFunction(mod, "pg_query_parse_plpgsql"),
		fPgQueryFreePlpgsqlParseResult:  newLazyFunction(mod, "pg_query_free_plpgsql_parse_result"),
		fPgQueryScan:                    newLazyFunction(mod, "pg_query_scan"),
		fPgQueryFreeScanResult:          newLazyFunction(mod, "pg_query_free_scan_result"),
		fPgQueryNormalize:               newLazyFunction(mod, "pg_query_normalize"),
		fPgQueryFreeNormalizeResult:     newLazyFunction(mod, "pg_query_free_normalize_result"),
		fPgQueryFingerprint:             newLazyFunction(mod, "pg_query_fingerprint"),
		fPgQueryFreeFingerprintResult:   newLazyFunction(mod, "pg_query_free_fingerprint_result"),
		fPgQueryDeparseProtobuf:         newLazyFunction(mod, "pg_query_deparse_protobuf"),
		fPgQueryFreeDeparseResult:       newLazyFunction(mod, "pg_query_free_deparse_result"),
		hashXXH364:                      newLazyFunction(mod, "XXH3_64bits_withSeed"),

		malloc: newLazyFunction(mod, "malloc"),
		free:   newLazyFunction(mod, "free"),

		mod:        mod,
		wasmMemory: mod.Memory(),
	}

	if err := res.pgQueryInit(); err != nil {
		res.discard()
		panic(err)
	}

	return res
}

func getABI() *abi {
	wasmInitOnce.Do(initWASM)

	abiPoolMu.Lock()

	e := abiPool.Front()
	if e == nil {
		abiPoolMu.Unlock()
		return newABI()
	}

	abiPool.Remove(e)
	abiPoolMu.Unlock()
	return e.Value.(*abi) //nolint:forcetypeassert // Private implementation detail: only *abi values are stored in abiPool.
}

func abiPoolLen() int {
	abiPoolMu.Lock()
	defer abiPoolMu.Unlock()
	return abiPool.Len()
}

type abi struct {
	fPgQueryInit                    lazyFunction
	fPgQueryParse                   lazyFunction
	fPgQueryFreeParseResult         lazyFunction
	fPgQueryParseProtobuf           lazyFunction
	fPgQueryFreeProtobufParseResult lazyFunction
	fPgQueryParsePlpgsql            lazyFunction
	fPgQueryFreePlpgsqlParseResult  lazyFunction
	fPgQueryScan                    lazyFunction
	fPgQueryFreeScanResult          lazyFunction
	fPgQueryNormalize               lazyFunction
	fPgQueryFreeNormalizeResult     lazyFunction
	fPgQueryFingerprint             lazyFunction
	fPgQueryFreeFingerprintResult   lazyFunction
	fPgQueryDeparseProtobuf         lazyFunction
	fPgQueryFreeDeparseResult       lazyFunction
	hashXXH364                      lazyFunction

	malloc lazyFunction
	free   lazyFunction

	wasmMemory api.Memory

	mod api.Module
}

func (abi *abi) Close() {
	abiPoolMu.Lock()
	abiPool.PushBack(abi)
	abiPoolMu.Unlock()
}

// discard drops the ABI without returning it to abiPool.
// Child modules share the root module's linear memory; closing one can fault
// the process, so we intentionally leak the instance. This path is rare
// (WASM/runtime failures and panics only).
func (abi *abi) discard() {
	// Do not Close the module and do not push onto abiPool.
}

func (abi *abi) pgQueryInit() error {
	_, err := abi.fPgQueryInit.Call0(context.Background())
	return err
}

func (abi *abi) pgQueryParse(input cString) (result string, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 12)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryParse.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return "", err
	}
	defer func() { _, _ = abi.fPgQueryFreeParseResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 12)
	if !ok {
		return "", errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[8:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	result, err = readCStringPtr(abi.wasmMemory, uint32(resPtr))
	return result, err
}

func (abi *abi) pgQueryParseProtobuf(input cString) (result []byte, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 16)
	if err != nil {
		return nil, err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryParseProtobuf.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return nil, err
	}
	defer func() { _, _ = abi.fPgQueryFreeProtobufParseResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 16)
	if !ok {
		return nil, errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[12:])
	if errPtr != 0 {
		return nil, newPgQueryError(abi.mod, errPtr)
	}

	pgQueryProtobufLen := binary.LittleEndian.Uint32(resBuf)
	pgQueryProtobufData := binary.LittleEndian.Uint32(resBuf[4:])

	buf, ok := abi.wasmMemory.Read(pgQueryProtobufData, pgQueryProtobufLen)
	if !ok {
		return nil, errFailedRead
	}

	result = bytes.Clone(buf)

	return result, nil
}

func (abi *abi) pgQueryDeParseFromProtobuf(input cString) (result string, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 8)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	paramPtr, err := abi.malloc.Call1(ctx, 8)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, paramPtr) }()

	abi.wasmMemory.WriteUint32Le(uint32(paramPtr), uint32(input.length))
	abi.wasmMemory.WriteUint32Le(uint32(paramPtr+4), input.ptr)

	if _, err := abi.fPgQueryDeparseProtobuf.Call2(ctx, resPtr, paramPtr); err != nil {
		return "", err
	}
	defer func() { _, _ = abi.fPgQueryFreeDeparseResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 8)
	if !ok {
		return "", errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[4:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	return readCStringPtr(abi.wasmMemory, uint32(resPtr))
}

func (abi *abi) pgQueryScanProtobuf(input cString) (result []byte, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 16)
	if err != nil {
		return nil, err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryScan.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return nil, err
	}
	defer func() { _, _ = abi.fPgQueryFreeScanResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 16)
	if !ok {
		return nil, errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[12:])
	if errPtr != 0 {
		return nil, newPgQueryError(abi.mod, errPtr)
	}

	pgQueryProtobufLen := binary.LittleEndian.Uint32(resBuf)
	pgQueryProtobufData := binary.LittleEndian.Uint32(resBuf[4:])

	buf, ok := abi.wasmMemory.Read(pgQueryProtobufData, pgQueryProtobufLen)
	if !ok {
		return nil, errFailedRead
	}

	result = bytes.Clone(buf)

	return result, nil
}

func (abi *abi) pgQueryNormalize(input cString) (result string, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 8)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryNormalize.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return "", err
	}
	defer func() { _, _ = abi.fPgQueryFreeNormalizeResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 8)
	if !ok {
		return "", errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[4:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	return readCStringPtr(abi.wasmMemory, uint32(resPtr))
}

func (abi *abi) pgQueryParsePlPgSqlToJSON(input cString) (result string, err error) { //nolint:revive // Match upstream method name
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 8)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryParsePlpgsql.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return "", err
	}
	defer func() { _, _ = abi.fPgQueryFreePlpgsqlParseResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 8)
	if !ok {
		return "", errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[4:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	return readCStringPtr(abi.wasmMemory, uint32(resPtr))
}

func (abi *abi) pgQueryFingerprintToUint64(input cString) (result uint64, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 20)
	if err != nil {
		return 0, err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryFingerprint.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return 0, err
	}
	defer func() { _, _ = abi.fPgQueryFreeFingerprintResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 20)
	if !ok {
		return 0, errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[16:])
	if errPtr != 0 {
		return 0, newPgQueryError(abi.mod, errPtr)
	}

	result = binary.LittleEndian.Uint64(resBuf)

	return result, nil
}

func (abi *abi) pgQueryFingerprintToHexStr(input cString) (result string, err error) {
	ctx := context.Background()

	resPtr, err := abi.malloc.Call1(ctx, 20)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = abi.free.Call1(ctx, resPtr) }()

	if _, err := abi.fPgQueryFingerprint.Call2(ctx, resPtr, uint64(input.ptr)); err != nil {
		return "", err
	}
	defer func() { _, _ = abi.fPgQueryFreeFingerprintResult.Call1(ctx, resPtr) }()

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 20)
	if !ok {
		return "", errFailedRead
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[16:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	return readCStringPtr(abi.wasmMemory, uint32(resPtr)+8)
}

func (abi *abi) pgQueryHashXXH364(input cString, seed uint64) (uint64, error) {
	ctx := context.Background()

	res, err := abi.hashXXH364.Call3(ctx, uint64(input.ptr), uint64(input.length), seed) //nolint:gosec // length is positive
	return res, err
}

func newPgQueryError(mod api.Module, errPtr uint32) error {
	message, err := readCStringPtr(mod.Memory(), errPtr)
	if err != nil {
		return err
	}
	funcname, err := readCStringPtr(mod.Memory(), errPtr+4)
	if err != nil {
		return err
	}
	filename, err := readCStringPtr(mod.Memory(), errPtr+8)
	if err != nil {
		return err
	}
	lineno, ok := mod.Memory().ReadUint32Le(errPtr + 12)
	if !ok {
		return errFailedRead
	}
	cursorpos, ok := mod.Memory().ReadUint32Le(errPtr + 16)
	if !ok {
		return errFailedRead
	}
	context, err := readCStringPtr(mod.Memory(), errPtr+20)
	if err != nil {
		return err
	}

	return &Error{
		Message:   message,
		Funcname:  funcname,
		Filename:  filename,
		Lineno:    int(lineno),
		Cursorpos: int(cursorpos),
		Context:   context,
	}
}

func readCStringPtr(mem api.Memory, ptrptr uint32) (string, error) {
	ptr, ok := mem.ReadUint32Le(ptrptr)
	if !ok {
		return "", errFailedRead
	}
	s := ""
	if ptr == 0 {
		return s, nil
	}
	endPtr := ptr
	for {
		b, ok := mem.ReadByte(endPtr)
		if !ok {
			return "", errFailedRead
		}
		if b == 0 {
			break
		}
		endPtr++
	}
	buf, ok := mem.Read(ptr, endPtr-ptr)
	if !ok {
		return "", errFailedRead
	}
	return string(buf), nil
}

type lazyFunction struct {
	f    api.Function
	name string
	mod  api.Module
}

func newLazyFunction(mod api.Module, name string) lazyFunction {
	return lazyFunction{mod: mod, name: name}
}

func (f *lazyFunction) Call0(ctx context.Context) (uint64, error) {
	var callStack [1]uint64
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call1(ctx context.Context, arg1 uint64) (uint64, error) {
	var callStack [1]uint64
	callStack[0] = arg1
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call2(ctx context.Context, arg1 uint64, arg2 uint64) (uint64, error) {
	var callStack [2]uint64
	callStack[0] = arg1
	callStack[1] = arg2
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call3(ctx context.Context, arg1 uint64, arg2 uint64, arg3 uint64) (uint64, error) {
	var callStack [3]uint64
	callStack[0] = arg1
	callStack[1] = arg2
	callStack[2] = arg3
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) callWithStack(ctx context.Context, callStack []uint64) (uint64, error) {
	if f.f == nil {
		f.f = f.mod.ExportedFunction(f.name)
		if f.f == nil {
			return 0, fmt.Errorf("go-pgquery: exported function %q not found", f.name)
		}
	}
	if err := f.f.CallWithStack(ctx, callStack); err != nil {
		return 0, err
	}
	return callStack[0], nil
}

type cString struct {
	ptr    uint32
	length int
	abi    *abi
}

func (abi *abi) newCString(s string) (cString, error) {
	ptr, err := abi.malloc.Call1(context.Background(), uint64(len(s)+1))
	if err != nil {
		return cString{}, err
	}
	if !abi.wasmMemory.WriteString(uint32(ptr), s) {
		_, _ = abi.free.Call1(context.Background(), ptr)
		return cString{}, errFailedWrite
	}
	if !abi.wasmMemory.WriteByte(uint32(ptr)+uint32(len(s)), 0) { //nolint:gosec // string must fit in 32-bit
		_, _ = abi.free.Call1(context.Background(), ptr)
		return cString{}, errFailedWrite
	}
	return cString{
		ptr:    uint32(ptr),
		length: len(s),
		abi:    abi,
	}, nil
}

func (abi *abi) newCStringFromBytes(s []byte) (cString, error) {
	ptr, err := abi.malloc.Call1(context.Background(), uint64(len(s)+1))
	if err != nil {
		return cString{}, err
	}
	if !abi.wasmMemory.Write(uint32(ptr), s) {
		_, _ = abi.free.Call1(context.Background(), ptr)
		return cString{}, errFailedWrite
	}
	if !abi.wasmMemory.WriteByte(uint32(ptr)+uint32(len(s)), 0) { //nolint:gosec // string must fit in 32-bit
		_, _ = abi.free.Call1(context.Background(), ptr)
		return cString{}, errFailedWrite
	}
	return cString{
		ptr:    uint32(ptr),
		length: len(s),
		abi:    abi,
	}, nil
}

func (s cString) Close() {
	_, _ = s.abi.free.Call1(context.Background(), uint64(s.ptr))
}
