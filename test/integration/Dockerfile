# Wrapper image that boots a real MikroTik RouterOS CHR in QEMU inside an
# Alpine container, used by the routeros Go client integration tests.
# REST (guest tcp/80) is forwarded to container tcp/80 via user-mode (SLIRP)
# networking; no --privileged, no NET_ADMIN, no bridge/tap required.
#
# Build the host-native variant (arch defaults to the platform being built,
# e.g. aarch64 on Apple Silicon / amd64 on x86 hosts):
#   docker build -t routeros-it-chr:test test/integration
# Build an explicit arch variant (Go-style and qemu-style names both work):
#   docker build --build-arg ROUTEROS_ARCH=aarch64 -t routeros-it-chr:test test/integration
#   docker build --build-arg ROUTEROS_ARCH=x86_64  -t routeros-it-chr:test test/integration

# x86_64 or aarch64; defaults to the platform being built (TARGETARCH is
# provided automatically by BuildKit/buildx per target platform) so a
# multi-platform build produces arch-matched emulators and CHR disks;
# an explicit --build-arg ROUTEROS_ARCH still overrides.
ARG TARGETARCH
ARG ROUTEROS_ARCH=${TARGETARCH}

# The arch normalization stage aliases: buildx's automatic TARGETARCH hands
# out Go-style names (amd64/arm64) while the qemu packages, the CHR assets
# and the entrypoint speak x86_64/aarch64, so both spellings are accepted on
# input and mapped onto the qemu names by the arch-${ROUTEROS_ARCH} base
# selection below. This is the one BuildKit mechanism that bakes the
# NORMALIZED value into the image config as the ROUTEROS_ARCH env var (the
# entrypoint picks the emulator binary from it): a `case` mapping inside a
# RUN block cannot do it, because shell assignments never persist into
# later Dockerfile instructions (ENV would see the raw input again). An
# unsupported ROUTEROS_ARCH value fails the build right at the arch-<name>
# stage lookup.
FROM alpine:3.21 AS arch-x86_64
ENV ROUTEROS_ARCH=x86_64

FROM alpine:3.21 AS arch-amd64
ENV ROUTEROS_ARCH=x86_64

FROM alpine:3.21 AS arch-aarch64
ENV ROUTEROS_ARCH=aarch64

FROM alpine:3.21 AS arch-arm64
ENV ROUTEROS_ARCH=aarch64

FROM arch-${ROUTEROS_ARCH}

ARG ROUTEROS_VERSION=7.23.3

# Only the QEMU system emulator matching ROUTEROS_ARCH is installed (a single
# interpolated package name: qemu-system-x86_64 OR qemu-system-aarch64), plus
# qemu-img for the VDI->qcow2 conversion and unzip for the CHR archive.
# busybox wget (bundled with alpine) handles the downloads.
RUN apk add --no-cache "qemu-system-${ROUTEROS_ARCH}" qemu-img unzip

WORKDIR /chr

# Download the CHR disk archive, trying download.mikrotik.com first and
# cdn.mikrotik.com as fallback (same path pattern). Convert the VDI to qcow2
# and delete the archive and VDI in the same layer so the image does not
# carry the ~230MB of intermediate files. Failing hard on a missing asset
# beats silently building an image that cannot boot the requested arch.
RUN set -eu; \
    case "${ROUTEROS_ARCH}" in \
        x86_64)  asset="chr-${ROUTEROS_VERSION}.vdi.zip" ;; \
        aarch64) asset="chr-${ROUTEROS_VERSION}-arm64.vdi.zip" ;; \
        *) echo "ERROR: unsupported ROUTEROS_ARCH '${ROUTEROS_ARCH}' (use x86_64 or aarch64)" >&2; exit 1 ;; \
    esac; \
    ok=""; \
    for host in https://download.mikrotik.com https://cdn.mikrotik.com; do \
        url="${host}/routeros/${ROUTEROS_VERSION}/${asset}"; \
        echo "downloading ${url}"; \
        if wget -q -T 120 "${url}" -O "${asset}"; then ok="1"; break; fi; \
        echo "WARN: download failed from ${url}" >&2; \
        rm -f "${asset}"; \
    done; \
    if [ -z "${ok}" ]; then \
        echo "ERROR: failed to download ${asset} for ${ROUTEROS_ARCH} from both download.mikrotik.com and cdn.mikrotik.com" >&2; \
        exit 1; \
    fi; \
    unzip -q "${asset}"; \
    rm -f "${asset}"; \
    vdi=$(ls chr-*.vdi 2>/dev/null | head -n 1 || true); \
    if [ -z "${vdi}" ]; then \
        echo "ERROR: no chr-*.vdi found inside ${asset}" >&2; \
        exit 1; \
    fi; \
    qemu-img convert -f vdi -O qcow2 "${vdi}" /chr.qcow2; \
    rm -f "${vdi}"

COPY --chmod=755 entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
