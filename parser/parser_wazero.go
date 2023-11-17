//go:build !tinygo.wasm && !pgquery_cgo

package parser

import (
	"context"
	"encoding/binary"
	"errors"
	"os"

	"github.com/wasilibs/go-pgquery/internal/wasix_32v1"
	"github.com/wasilibs/go-pgquery/internal/wasm"
	wazero "github.com/wasilibs/wazerox"
	"github.com/wasilibs/wazerox/api"
	"github.com/wasilibs/wazerox/experimental"
	"github.com/wasilibs/wazerox/imports/wasi_snapshot_preview1"
)

var (
	errFailedWrite = errors.New("failed to write to wasm memory")
	errFailedRead  = errors.New("failed to read from wasm memory")
)

var (
	wasmRT       wazero.Runtime
	wasmCompiled wazero.CompiledModule
)

func init() {
	ctx := context.Background()
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2|experimental.CoreFeaturesThreads))

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	wasix_32v1.MustInstantiate(ctx, rt)

	code, err := rt.CompileModule(ctx, wasm.LibPGQuery)
	if err != nil {
		panic(err)
	}

	wasmCompiled = code
	wasmRT = rt
}

func newABI() *abi {
	cfg := wazero.NewModuleConfig().WithSysNanotime().WithStdout(os.Stdout).WithStderr(os.Stderr).WithStartFunctions("_initialize")
	mod, err := wasmRT.InstantiateModule(context.Background(), wasmCompiled, cfg)
	if err != nil {
		panic(err)
	}
	abi := &abi{
		fPgQueryInit:                    newLazyFunction(wasmRT, mod, "pg_query_init"),
		fPgQueryParse:                   newLazyFunction(wasmRT, mod, "pg_query_parse"),
		fPgQueryFreeParseResult:         newLazyFunction(wasmRT, mod, "pg_query_free_parse_result"),
		fPgQueryParseProtobuf:           newLazyFunction(wasmRT, mod, "pg_query_parse_protobuf"),
		fPgQueryFreeProtobufParseResult: newLazyFunction(wasmRT, mod, "pg_query_free_protobuf_parse_result"),
		fPgQueryParsePlpgsql:            newLazyFunction(wasmRT, mod, "pg_query_parse_plpgsql"),
		fPgQueryFreePlpgsqlParseResult:  newLazyFunction(wasmRT, mod, "pg_query_free_plpgsql_parse_result"),
		fPgQueryScan:                    newLazyFunction(wasmRT, mod, "pg_query_scan"),
		fPgQueryFreeScanResult:          newLazyFunction(wasmRT, mod, "pg_query_free_scan_result"),
		fPgQueryNormalize:               newLazyFunction(wasmRT, mod, "pg_query_normalize"),
		fPgQueryFreeNormalizeResult:     newLazyFunction(wasmRT, mod, "pg_query_free_normalize_result"),

		malloc: newLazyFunction(wasmRT, mod, "malloc"),
		free:   newLazyFunction(wasmRT, mod, "free"),

		mod:        mod,
		wasmMemory: mod.Memory(),
	}

	abi.pgQueryInit()

	return abi
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

	malloc lazyFunction
	free   lazyFunction

	wasmMemory api.Memory

	mod api.Module
}

func (abi *abi) Close() error {
	return abi.mod.Close(context.Background())
}

func (abi *abi) pgQueryInit() {
	abi.fPgQueryInit.Call0(context.Background())
}

func (abi *abi) pgQueryParse(input cString) (result string, err error) {
	ctx := wasix_32v1.BackgroundContext()

	resPtr := abi.malloc.Call1(ctx, 12)
	defer abi.free.Call1(ctx, resPtr)

	abi.fPgQueryParse.Call2(ctx, resPtr, uint64(input.ptr))
	defer abi.fPgQueryFreeParseResult.Call1(ctx, resPtr)

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 12)
	if !ok {
		panic(errFailedRead)
	}
	parseTreePtr := binary.LittleEndian.Uint32(resBuf)
	parseTreeEndPtr := parseTreePtr
	for {
		if b, ok := abi.wasmMemory.ReadByte(parseTreeEndPtr); !ok {
			panic(errFailedRead)
		} else if b == 0 {
			break
		}
		parseTreeEndPtr++
	}

	buf, ok := abi.wasmMemory.Read(parseTreePtr, parseTreeEndPtr-parseTreePtr)
	if !ok {
		panic(errFailedRead)
	}

	return string(buf), nil
}

func (abi *abi) pgQueryParseProtobuf(input cString) (result []byte, err error) {
	ctx := wasix_32v1.BackgroundContext()

	resPtr := abi.malloc.Call1(ctx, 16)
	defer abi.free.Call1(ctx, resPtr)

	abi.fPgQueryParseProtobuf.Call2(ctx, resPtr, uint64(input.ptr))
	defer abi.fPgQueryFreeProtobufParseResult.Call1(ctx, resPtr)

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 16)
	if !ok {
		panic(errFailedRead)
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[12:])
	if errPtr != 0 {
		return nil, newPgQueryError(abi.mod, errPtr)
	}

	pgQueryProtobufLen := binary.LittleEndian.Uint32(resBuf)
	pgQueryProtobufData := binary.LittleEndian.Uint32(resBuf[4:])

	buf, ok := abi.wasmMemory.Read(pgQueryProtobufData, pgQueryProtobufLen)
	if !ok {
		panic(errFailedRead)
	}

	return buf, nil
}

func (abi *abi) pgQueryScanProtobuf(input cString) (result []byte, err error) {
	ctx := wasix_32v1.BackgroundContext()

	resPtr := abi.malloc.Call1(ctx, 16)
	defer abi.free.Call1(ctx, resPtr)

	abi.fPgQueryScan.Call2(ctx, resPtr, uint64(input.ptr))
	defer abi.fPgQueryFreeScanResult.Call1(ctx, resPtr)

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 16)
	if !ok {
		panic(errFailedRead)
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[12:])
	if errPtr != 0 {
		return nil, newPgQueryError(abi.mod, errPtr)
	}

	pgQueryProtobufLen := binary.LittleEndian.Uint32(resBuf)
	pgQueryProtobufData := binary.LittleEndian.Uint32(resBuf[4:])

	buf, ok := abi.wasmMemory.Read(pgQueryProtobufData, pgQueryProtobufLen)
	if !ok {
		panic(errFailedRead)
	}

	return buf, nil
}

func (abi *abi) pgQueryNormalize(input cString) (result string, err error) {
	ctx := wasix_32v1.BackgroundContext()

	resPtr := abi.malloc.Call1(ctx, 8)
	defer abi.free.Call1(ctx, resPtr)

	abi.fPgQueryNormalize.Call2(ctx, resPtr, uint64(input.ptr))
	defer abi.fPgQueryFreeNormalizeResult.Call1(ctx, resPtr)

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 8)
	if !ok {
		panic(errFailedRead)
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[4:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	result = readCStringPtr(abi.wasmMemory, uint32(resPtr))

	return
}

func (abi *abi) pgQueryParsePlPgSqlToJSON(input cString) (result string, err error) {
	ctx := wasix_32v1.BackgroundContext()

	resPtr := abi.malloc.Call1(ctx, 8)
	defer abi.free.Call1(ctx, resPtr)

	abi.fPgQueryParsePlpgsql.Call2(ctx, resPtr, uint64(input.ptr))
	defer abi.fPgQueryFreePlpgsqlParseResult.Call1(ctx, resPtr)

	resBuf, ok := abi.wasmMemory.Read(uint32(resPtr), 8)
	if !ok {
		panic(errFailedRead)
	}

	errPtr := binary.LittleEndian.Uint32(resBuf[4:])
	if errPtr != 0 {
		return "", newPgQueryError(abi.mod, errPtr)
	}

	result = readCStringPtr(abi.wasmMemory, uint32(resPtr))

	return
}

func newPgQueryError(mod api.Module, errPtr uint32) error {
	message := readCStringPtr(mod.Memory(), errPtr)
	funcname := readCStringPtr(mod.Memory(), errPtr+4)
	filename := readCStringPtr(mod.Memory(), errPtr+8)
	lineno, ok := mod.Memory().ReadUint32Le(errPtr + 12)
	if !ok {
		panic(errFailedRead)
	}
	cursorpos, ok := mod.Memory().ReadUint32Le(errPtr + 16)
	if !ok {
		panic(errFailedRead)
	}
	context := readCStringPtr(mod.Memory(), errPtr+20)

	return &Error{
		Message:   message,
		Funcname:  funcname,
		Filename:  filename,
		Lineno:    int(lineno),
		Cursorpos: int(cursorpos),
		Context:   context,
	}
}

func readCStringPtr(mem api.Memory, ptrptr uint32) string {
	ptr, ok := mem.ReadUint32Le(ptrptr)
	if !ok {
		panic(errFailedRead)
	}
	s := ""
	if ptr == 0 {
		return s
	}
	endPtr := ptr
	for {
		if b, ok := mem.ReadByte(endPtr); !ok {
			panic(errFailedRead)
		} else if b == 0 {
			break
		}
		endPtr++
	}
	buf, ok := mem.Read(ptr, endPtr-ptr)
	if !ok {
		panic(errFailedRead)
	}
	return string(buf)
}

type lazyFunction struct {
	f    api.Function
	rt   wazero.Runtime
	name string
	mod  api.Module
}

func newLazyFunction(rt wazero.Runtime, mod api.Module, name string) lazyFunction {
	return lazyFunction{rt: rt, mod: mod, name: name}
}

func (f *lazyFunction) Call0(ctx context.Context) uint64 {
	var callStack [1]uint64
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call1(ctx context.Context, arg1 uint64) uint64 {
	var callStack [1]uint64
	callStack[0] = arg1
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call2(ctx context.Context, arg1 uint64, arg2 uint64) uint64 {
	var callStack [2]uint64
	callStack[0] = arg1
	callStack[1] = arg2
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call3(ctx context.Context, arg1 uint64, arg2 uint64, arg3 uint64) uint64 {
	var callStack [3]uint64
	callStack[0] = arg1
	callStack[1] = arg2
	callStack[2] = arg3
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) Call8(ctx context.Context, arg1 uint64, arg2 uint64, arg3 uint64, arg4 uint64, arg5 uint64, arg6 uint64, arg7 uint64, arg8 uint64) uint64 {
	var callStack [8]uint64
	callStack[0] = arg1
	callStack[1] = arg2
	callStack[2] = arg3
	callStack[3] = arg4
	callStack[4] = arg5
	callStack[5] = arg6
	callStack[6] = arg7
	callStack[7] = arg8
	return f.callWithStack(ctx, callStack[:])
}

func (f *lazyFunction) callWithStack(ctx context.Context, callStack []uint64) uint64 {
	if f.f == nil {
		f.f = f.mod.ExportedFunction(f.name)
	}
	if err := f.f.CallWithStack(ctx, callStack); err != nil {
		panic(err)
	}
	return callStack[0]
}

type cString struct {
	ptr    uint32
	length int
	abi    *abi
}

func (abi *abi) newCString(s string) cString {
	ptr := uint32(abi.malloc.Call1(context.Background(), uint64(len(s)+1)))
	if !abi.wasmMemory.WriteString(ptr, s) {
		panic(errFailedWrite)
	}
	if !abi.wasmMemory.WriteByte(ptr+uint32(len(s)), 0) {
		panic(errFailedWrite)
	}
	return cString{
		ptr:    ptr,
		length: len(s),
		abi:    abi,
	}
}

func (s cString) Close() error {
	s.abi.free.Call1(context.Background(), uint64(s.ptr))
	return nil
}
