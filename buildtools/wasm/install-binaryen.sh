#!/bin/sh
set -eux
binaryen_target_arch="${TARGETARCH:-$(uname -m)}"
case "${binaryen_target_arch}" in
  amd64)
    binaryen_arch=x86_64-linux
    binaryen_sha256="${BINARYEN_X86_64_SHA256}"
    ;;
  x86_64)
    binaryen_arch=x86_64-linux
    binaryen_sha256="${BINARYEN_X86_64_SHA256}"
    ;;
  arm64)
    binaryen_arch=aarch64-linux
    binaryen_sha256="${BINARYEN_AARCH64_SHA256}"
    ;;
  aarch64)
    binaryen_arch=aarch64-linux
    binaryen_sha256="${BINARYEN_AARCH64_SHA256}"
    ;;
  *)
    echo "unsupported TARGETARCH: ${binaryen_target_arch}"
    exit 1
    ;;
esac
binaryen_tar="binaryen-${BINARYEN_VERSION}-${binaryen_arch}.tar.gz"
curl -fsSL -o "/tmp/${binaryen_tar}" \
  "https://github.com/WebAssembly/binaryen/releases/download/${BINARYEN_VERSION}/${binaryen_tar}"
echo "${binaryen_sha256}  /tmp/${binaryen_tar}" | sha256sum -c -
tar -xzf "/tmp/${binaryen_tar}" -C /opt
mv "/opt/binaryen-${BINARYEN_VERSION}" /opt/binaryen
rm "/tmp/${binaryen_tar}"
