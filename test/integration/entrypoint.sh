#!/bin/sh
# Boots the bundled RouterOS CHR disk in QEMU with user-mode (SLIRP)
# networking and REST (guest tcp/80) forwarded to container tcp/80.
# QEMU replaces this shell as PID 1 and stays in the foreground so Docker
# considers the container alive for as long as the router runs.
set -eu

# RouterOS resources are tunable through the container environment:
# QEMU_MEMORY (MiB) and QEMU_CPUS (vCPU count). Each knob must be a
# positive integer; an invalid or missing value WARNs to stderr and falls
# back to the default so a malformed knob never fails the boot.
QEMU_MEMORY="${QEMU_MEMORY:-512}"
QEMU_CPUS="${QEMU_CPUS:-2}"

case "$QEMU_MEMORY" in
    *[!0-9]*|'')
        echo "WARN: QEMU_MEMORY='${QEMU_MEMORY}' is not a positive integer, using default 512" >&2
        QEMU_MEMORY=512
        ;;
    *)
        if [ "$QEMU_MEMORY" -lt 1 ]; then
            echo "WARN: QEMU_MEMORY='${QEMU_MEMORY}' is not a positive integer, using default 512" >&2
            QEMU_MEMORY=512
        fi
        ;;
esac

case "$QEMU_CPUS" in
    *[!0-9]*|'')
        echo "WARN: QEMU_CPUS='${QEMU_CPUS}' is not a positive integer, using default 2" >&2
        QEMU_CPUS=2
        ;;
    *)
        if [ "$QEMU_CPUS" -lt 1 ]; then
            echo "WARN: QEMU_CPUS='${QEMU_CPUS}' is not a positive integer, using default 2" >&2
            QEMU_CPUS=2
        fi
        ;;
esac

QEMU="/usr/bin/qemu-system-${ROUTEROS_ARCH}"

case "${ROUTEROS_ARCH}" in
    x86_64)
        # SeaBIOS (built into QEMU) boots the CHR disk directly and
        # if=virtio maps to PCI on the q35 machine.
        set -- \
            -M q35 \
            -drive file=/chr.qcow2,format=qcow2,if=virtio
        ;;
    aarch64)
        # The virt board boots CHR only through UEFI firmware: the edk2
        # code pflash ships with the qemu-system-aarch64 package, and a
        # writable vars pflash of the same size must sit next to it
        # (edk2 initializes the empty varstore on first boot). if=virtio
        # maps to MMIO on the virt board and CHR never finds the disk, so
        # the disk is attached through an explicit virtio-blk-pci device.
        code_fd=/usr/share/qemu/edk2-aarch64-code.fd
        vars_fd=/chr-efi-vars.fd
        head -c "$(stat -c %s "${code_fd}")" /dev/zero >"${vars_fd}"
        set -- \
            -M virt \
            -drive if=pflash,format=raw,file="${code_fd}",readonly=on \
            -drive if=pflash,format=raw,file="${vars_fd}" \
            -drive file=/chr.qcow2,format=qcow2,if=none,id=chr0 \
            -device virtio-blk-pci,drive=chr0
        ;;
    *)
        echo "entrypoint: unsupported ROUTEROS_ARCH '${ROUTEROS_ARCH}'" >&2
        exit 1
        ;;
esac

# QEMU does not fall back on its own when -enable-kvm fails, so the
# acceleration flags are conditional on /dev/kvm being writable. Without
# it (macOS Docker Desktop, most CI runners) the router boots under TCG
# software emulation, which is expected and must work.
if [ -w /dev/kvm ]; then
    set -- "$@" -enable-kvm -cpu host
elif [ "${ROUTEROS_ARCH}" = "aarch64" ]; then
    # CHR arm64 images carry a 32-bit ARM init; cortex-a710 is the proven
    # TCG CPU model for them (see tikoci/quickchr qemu instructions).
    set -- "$@" -cpu cortex-a710
else
    set -- "$@" -cpu max
fi

exec "$QEMU" "$@" \
    -m "$QEMU_MEMORY" \
    -smp "$QEMU_CPUS" \
    -netdev user,id=net0,hostfwd=tcp::80-:80 \
    -device virtio-net-pci,netdev=net0 \
    -display none \
    -serial none \
    -monitor none
