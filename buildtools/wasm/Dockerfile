FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-32

ARG TARGETARCH
ARG BINARYEN_VERSION=version_125
ARG BINARYEN_X86_64_SHA256=7c3bc16599c8274a04d34a504fe4be2047884f900e0e2da2f6fb9cd667183be4
ARG BINARYEN_AARCH64_SHA256=d0382de3c189a7cbb9fdda93e2966f4557f6a00f201e2c4937c27ca01cead4fc

RUN apt-get update && apt-get install -y ca-certificates curl patch && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    binaryen_target_arch="${TARGETARCH:-$(uname -m)}"; \
    case "${binaryen_target_arch}" in \
      amd64) binaryen_arch=x86_64-linux; binaryen_sha256="${BINARYEN_X86_64_SHA256}" ;; \
      x86_64) binaryen_arch=x86_64-linux; binaryen_sha256="${BINARYEN_X86_64_SHA256}" ;; \
      arm64) binaryen_arch=aarch64-linux; binaryen_sha256="${BINARYEN_AARCH64_SHA256}" ;; \
      aarch64) binaryen_arch=aarch64-linux; binaryen_sha256="${BINARYEN_AARCH64_SHA256}" ;; \
      *) echo "unsupported TARGETARCH: ${binaryen_target_arch}"; exit 1 ;; \
    esac; \
    binaryen_tar="binaryen-${BINARYEN_VERSION}-${binaryen_arch}.tar.gz"; \
    curl -fsSL -o "/tmp/${binaryen_tar}" \
      "https://github.com/WebAssembly/binaryen/releases/download/${BINARYEN_VERSION}/${binaryen_tar}"; \
    echo "${binaryen_sha256}  /tmp/${binaryen_tar}" | sha256sum -c -; \
    tar -xzf "/tmp/${binaryen_tar}" -C /opt; \
    mv "/opt/binaryen-${BINARYEN_VERSION}" /opt/binaryen; \
    rm "/tmp/${binaryen_tar}"

ENV PATH="/opt/binaryen/bin:${PATH}"

ENV CFLAGS="-pthread --target=wasm32-wasip1-threads --sysroot=/opt/wasi-sdk/share/wasi-sysroot -mllvm -wasm-enable-sjlj -mllvm -wasm-use-legacy-eh=false -O3 -pthread -D_WASI_EMULATED_PROCESS_CLOCKS -D_WASI_EMULATED_MMAN -D_WASI_EMULATED_SIGNAL -DPLATFORM_DEFAULT_WAL_SYNC_METHOD=WAL_SYNC_METHOD_OPEN_DSYNC"

WORKDIR /workspace
ADD internal/cparser .
ADD buildtools/wasm/patch.txt patch.txt
RUN patch -p3 < patch.txt

RUN CMD="$CC -Iinclude -Iinclude/postgres -fstack-protector -std=gnu99 -Wno-unknown-warning-option -c $CFLAGS *.c" && echo $CMD && $CMD
RUN $AR rs libpg_query.a *.o

ENV LDFLAGS="--target=wasm32-wasip1-threads --sysroot=/opt/wasi-sdk/share/wasi-sysroot -mexec-model=reactor -pthread -lsetjmp -lwasi-emulated-process-clocks -lwasi-emulated-mman -lwasi-emulated-signal -Wl,--shared-memory -Wl,--max-memory=4294967296"

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
  -Wl,--export=__stack_pointer \
  -Wl,--export=__heap_base

RUN wasm-opt -o libpg_query.so --translate-to-exnref -O3 libpg_query-noopt.so

CMD ["bash", "-c", "cp libpg_query.so /out/"]
