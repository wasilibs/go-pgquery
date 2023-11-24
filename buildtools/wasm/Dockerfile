FROM ghcr.io/wasilibs/wasix-sdk:sha-fc94d60

RUN apt-get update && apt-get install -y binaryen curl patch

ENV CFLAGS "${CFLAGS} -O3 -D_WASI_EMULATED_PROCESS_CLOCKS -D_WASI_EMULATED_MMAN"
ENV LDFLAGS $LDFLAGS -mexec-model=reactor -lwasi-emulated-process-clocks -lwasi-emulated-mman

WORKDIR /workspace
RUN curl -L https://github.com/pganalyze/libpg_query/archive/refs/tags/15-4.2.3.tar.gz | tar -xz --strip-components 1 -C /workspace
ADD buildtools/wasm/patch.txt patch.txt
RUN patch -p1 < patch.txt

RUN make build && \
    $RANLIB libpg_query.a

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
    -Wl,--export=__stack_pointer \
    -Wl,--export=__heap_base

RUN wasm-opt -o libpg_query.so --low-memory-unused --flatten --rereloop -O3 libpg_query-noopt.so

CMD ["bash", "-c", "cp libpg_query.so /out/"]
