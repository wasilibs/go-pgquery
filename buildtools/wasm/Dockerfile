FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-34

ARG TARGETARCH
ARG BINARYEN_VERSION=version_125
ARG BINARYEN_X86_64_SHA256=7c3bc16599c8274a04d34a504fe4be2047884f900e0e2da2f6fb9cd667183be4
ARG BINARYEN_AARCH64_SHA256=d0382de3c189a7cbb9fdda93e2966f4557f6a00f201e2c4937c27ca01cead4fc

RUN apt-get update && apt-get install -y ca-certificates curl patch wabt && rm -rf /var/lib/apt/lists/*

ADD buildtools/wasm/install-binaryen.sh /tmp/install-binaryen.sh
RUN ["/tmp/install-binaryen.sh"]

ENV PATH="/opt/binaryen/bin:${PATH}"

ENV CFLAGS="-pthread --target=wasm32-wasip1-threads --sysroot=/opt/wasi-sdk/share/wasi-sysroot -mllvm -wasm-enable-sjlj -mllvm -wasm-use-legacy-eh=false -O3 -pthread -D_WASI_EMULATED_PROCESS_CLOCKS -D_WASI_EMULATED_MMAN -D_WASI_EMULATED_SIGNAL -DPLATFORM_DEFAULT_WAL_SYNC_METHOD=WAL_SYNC_METHOD_OPEN_DSYNC"

WORKDIR /workspace
ADD internal/cparser .
ADD buildtools/wasm/patch.txt patch.txt
RUN patch -p3 < patch.txt

RUN CMD="$CC -Iinclude -Iinclude/postgres -fstack-protector -std=gnu99 -Wno-unknown-warning-option -c $CFLAGS *.c" && echo $CMD && $CMD
RUN $AR rs libpg_query.a *.o

ENV LDFLAGS="--target=wasm32-wasip1-threads --sysroot=/opt/wasi-sdk/share/wasi-sysroot -mexec-model=reactor -pthread -lsetjmp -lwasi-emulated-process-clocks -lwasi-emulated-mman -lwasi-emulated-signal -Wl,--shared-memory -Wl,--import-memory -Wl,--export-memory -Wl,--max-memory=4294967296 -Wl,-z,stack-size=1048576"

RUN $CC -o libpg_query-noopt.so $LDFLAGS \
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
  -Wl,--export=pg_query_deparse_protobuf \
  -Wl,--export=pg_query_free_deparse_result \
  -Wl,--export=XXH3_64bits_withSeed \
  -Wl,--export=__wasm_init_tls \
  -Wl,--export=__stack_pointer \
  -Wl,--export=__tls_size \
  -Wl,--export=__tls_align

RUN wasm-opt -o libpg_query.so --translate-to-exnref -O3 libpg_query-noopt.so

RUN pages=$(wasm-objdump -x libpg_query.so | grep -oE 'initial=[0-9]+' | head -1 | grep -oE '[0-9]+') \
  && printf '(module (memory (export "memory") %s 65536 shared))' "$pages" > memory.wat \
  && wat2wasm --enable-threads memory.wat -o memory.wasm

CMD ["bash", "-c", "cp libpg_query.so memory.wasm /out/"]
