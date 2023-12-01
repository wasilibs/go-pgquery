FROM ghcr.io/wasilibs/wasix-sdk:sha-fc94d60

RUN apt-get update && apt-get install -y binaryen curl patch

ENV CFLAGS "${CFLAGS} -O3 -msimd128 -D_WASI_EMULATED_PROCESS_CLOCKS -D_WASI_EMULATED_MMAN"
ENV LDFLAGS $LDFLAGS -mexec-model=reactor -lwasi-emulated-process-clocks -lwasi-emulated-mman

WORKDIR /workspace
ADD internal/cparser .
ADD buildtools/wasm/patch.txt patch.txt
ADD buildtools/wasm/patch-musl.txt patch-musl.txt
ADD buildtools/wasm/patch-libpgquery.txt patch-libpgquery.txt
RUN patch -p3 < patch.txt
# Workaround bug in strtol in musl, used by wasi https://www.openwall.com/lists/musl/2023/12/01/1
RUN patch -p3 < patch-musl.txt
# Workaround obvious bug in libpgquery that seems to only cause problems with this toolchain
# https://github.com/pganalyze/libpg_query/pull/223
RUN patch -p3 < patch-libpgquery.txt

RUN CMD="$CC -Iinclude -fstack-protector -fno-strict-aliasing -fwrapv -fPIC -std=gnu99 -Wno-deprecated-non-prototype -c $CFLAGS *.c" && echo $CMD && $CMD
RUN $AR rs libpg_query.a *.o

RUN $CXX -o libpg_query-noopt.so -Wl,--global-base=1024 $LDFLAGS \
    libpg_query.a \
    -Wl,--export=malloc \
    -Wl,--export=free \
    -Wl,--export=pg_query_init \
    -Wl,--export=pg_query_parse \
    -Wl,--export=pg_query_free_parse_result \
    -Wl,--export=pg_query_parse_protobuf \
    -Wl,--export=pg_query_free_protobuf_parse_result \
    -Wl,--export=pg_query_parse_plpgsql \
    -Wl,--export=pg_query_free_plpgsql_parse_result \
    -Wl,--export=pg_query_scan \
    -Wl,--export=pg_query_free_scan_result \
    -Wl,--export=pg_query_normalize \
    -Wl,--export=pg_query_free_normalize_result \
    -Wl,--export=pg_query_fingerprint \
    -Wl,--export=pg_query_free_fingerprint_result \
    -Wl,--export=XXH3_64bits_withSeed \
    -Wl,--export=__stack_pointer \
    -Wl,--export=__heap_base

RUN wasm-opt -o libpg_query.so --low-memory-unused --flatten --rereloop -O3 libpg_query-noopt.so

CMD ["bash", "-c", "cp libpg_query.so /out/"]
