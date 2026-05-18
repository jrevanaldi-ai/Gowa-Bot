#!/usr/bin/env bash
# install.sh — installer interaktif untuk Gowa-Bot
# Usage:  ./install.sh         (interaktif)
#         ./install.sh --help
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_MIN_VERSION="1.26"
GO_INSTALL_VERSION="1.26.0"
BIN_NAME="gowa-bot"
ENV_FILE="${REPO_DIR}/.env"
ENV_EXAMPLE="${REPO_DIR}/.env.example"
SERVICE_FILE="/etc/systemd/system/${BIN_NAME}.service"

# ---------- pretty print ----------
if [ -t 1 ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]; then
    C_RESET="$(tput sgr0)"; C_BOLD="$(tput bold)"
    C_RED="$(tput setaf 1)"; C_GREEN="$(tput setaf 2)"; C_YELLOW="$(tput setaf 3)"
    C_BLUE="$(tput setaf 4)"; C_DIM="$(tput dim)"
else
    C_RESET=""; C_BOLD=""; C_RED=""; C_GREEN=""; C_YELLOW=""; C_BLUE=""; C_DIM=""
fi

info()    { printf "%s[*]%s %s\n" "$C_BLUE"   "$C_RESET" "$*"; }
ok()      { printf "%s[+]%s %s\n" "$C_GREEN"  "$C_RESET" "$*"; }
warn()    { printf "%s[!]%s %s\n" "$C_YELLOW" "$C_RESET" "$*"; }
err()     { printf "%s[x]%s %s\n" "$C_RED"    "$C_RESET" "$*" >&2; }
section() { printf "\n%s== %s ==%s\n" "$C_BOLD" "$*" "$C_RESET"; }

usage() {
    cat <<EOF
${C_BOLD}Gowa-Bot installer${C_RESET}

Usage: ./install.sh [options]

Options:
  --help          Tampilkan bantuan ini
  --skip-deps     Skip instalasi system deps (git/ffmpeg/curl)
  --skip-go       Skip pengecekan & instalasi Go
  --skip-env      Skip prompt isi .env (pakai .env.example apa adanya)
  --skip-build    Skip 'go build'
  --skip-service  Skip prompt systemd service
  --yes           Auto-confirm semua prompt non-isi (terima default)

Yang akan dilakukan:
  1. Cek OS & deteksi package manager (apt/dnf/pacman/brew)
  2. Install git, ffmpeg, curl (kalau belum ada)
  3. Install Go ${GO_INSTALL_VERSION} dari go.dev kalau versi < ${GO_MIN_VERSION}
  4. Bikin .env dari .env.example (interaktif)
  5. Build binary './${BIN_NAME}'
  6. Opsional: install systemd service
EOF
}

# ---------- args ----------
SKIP_DEPS=0; SKIP_GO=0; SKIP_ENV=0; SKIP_BUILD=0; SKIP_SERVICE=0; AUTO_YES=0
for arg in "$@"; do
    case "$arg" in
        -h|--help)       usage; exit 0 ;;
        --skip-deps)     SKIP_DEPS=1 ;;
        --skip-go)       SKIP_GO=1 ;;
        --skip-env)      SKIP_ENV=1 ;;
        --skip-build)    SKIP_BUILD=1 ;;
        --skip-service)  SKIP_SERVICE=1 ;;
        --yes|-y)        AUTO_YES=1 ;;
        *) err "Argumen tidak dikenal: $arg"; usage; exit 1 ;;
    esac
done

# ---------- helpers ----------
ask() {
    # ask "prompt" "default" -> echo answer
    local prompt="$1" default="${2:-}" answer
    if [ "$AUTO_YES" = 1 ]; then
        printf "%s\n" "$default"; return
    fi
    if [ -n "$default" ]; then
        read -r -p "$(printf "%s%s%s [%s]: " "$C_BOLD" "$prompt" "$C_RESET" "$default")" answer || true
        printf "%s\n" "${answer:-$default}"
    else
        read -r -p "$(printf "%s%s%s: " "$C_BOLD" "$prompt" "$C_RESET")" answer || true
        printf "%s\n" "$answer"
    fi
}

confirm() {
    # confirm "Question?" "Y" -> 0/1
    local prompt="$1" default="${2:-Y}" answer
    if [ "$AUTO_YES" = 1 ]; then return 0; fi
    local hint="[Y/n]"; [ "$default" = "N" ] && hint="[y/N]"
    read -r -p "$(printf "%s%s%s %s " "$C_BOLD" "$prompt" "$C_RESET" "$hint")" answer || true
    answer="${answer:-$default}"
    case "$answer" in [Yy]*) return 0 ;; *) return 1 ;; esac
}

have() { command -v "$1" >/dev/null 2>&1; }

sudo_run() {
    if [ "$(id -u)" -eq 0 ]; then
        "$@"
    elif have sudo; then
        sudo "$@"
    else
        err "Butuh root atau sudo untuk: $*"
        return 1
    fi
}

version_ge() {
    # version_ge 1.26.0 1.26 -> 0 (true)
    printf "%s\n%s\n" "$2" "$1" | sort -V -C
}

# ---------- OS detect ----------
detect_os() {
    OS_KIND="unknown"; PKG=""
    case "$(uname -s)" in
        Linux)
            OS_KIND="linux"
            if   have apt-get; then PKG="apt"
            elif have dnf;     then PKG="dnf"
            elif have pacman;  then PKG="pacman"
            elif have apk;     then PKG="apk"
            fi
            ;;
        Darwin)
            OS_KIND="darwin"
            have brew && PKG="brew"
            ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l|armv6l) ARCH="armv6l" ;;
        *) ARCH="$(uname -m)" ;;
    esac
}

pkg_install() {
    # pkg_install pkg1 pkg2 ...
    [ "$#" -eq 0 ] && return 0
    case "$PKG" in
        apt)    sudo_run apt-get update -qq && sudo_run apt-get install -y "$@" ;;
        dnf)    sudo_run dnf install -y "$@" ;;
        pacman) sudo_run pacman -Sy --noconfirm --needed "$@" ;;
        apk)    sudo_run apk add --no-cache "$@" ;;
        brew)   brew install "$@" ;;
        *)      err "Package manager tidak terdeteksi. Install manual: $*"; return 1 ;;
    esac
}

# ---------- step: system deps ----------
install_system_deps() {
    section "1. System dependencies"
    if [ "$SKIP_DEPS" = 1 ]; then warn "Skip (--skip-deps)"; return; fi

    local need=()
    have git    || need+=("git")
    have curl   || need+=("curl")
    have ffmpeg || need+=("ffmpeg")
    have tar    || need+=("tar")

    if [ "${#need[@]}" -eq 0 ]; then
        ok "Semua dependency sistem sudah terpasang (git, curl, ffmpeg, tar)"
        return
    fi

    info "Akan install: ${need[*]}"
    if [ -z "$PKG" ]; then
        err "Tidak ada package manager yang dikenali. Install manual: ${need[*]}"
        return 1
    fi
    confirm "Lanjut install via ${PKG}?" "Y" || { warn "Skip install deps"; return; }
    pkg_install "${need[@]}"
    ok "System deps terpasang"
}

# ---------- step: Go ----------
go_version() {
    have go || { echo ""; return; }
    go version 2>/dev/null | awk '{print $3}' | sed 's/^go//'
}

install_go() {
    section "2. Go toolchain (min ${GO_MIN_VERSION})"
    if [ "$SKIP_GO" = 1 ]; then warn "Skip (--skip-go)"; return; fi

    local cur; cur="$(go_version || true)"
    if [ -n "$cur" ] && version_ge "$cur" "$GO_MIN_VERSION"; then
        ok "Go ${cur} terpasang"
        return
    fi

    if [ -n "$cur" ]; then
        warn "Go ${cur} terlalu tua (butuh >= ${GO_MIN_VERSION})"
    else
        warn "Go belum terpasang"
    fi

    if [ "$OS_KIND" != "linux" ]; then
        err "Auto-install Go hanya didukung di Linux. Untuk macOS: 'brew install go' atau download manual dari https://go.dev/dl/"
        return 1
    fi

    local tarball="go${GO_INSTALL_VERSION}.linux-${ARCH}.tar.gz"
    local url="https://go.dev/dl/${tarball}"
    info "Download Go ${GO_INSTALL_VERSION} (${ARCH}) dari ${url}"
    confirm "Lanjut download & install ke /usr/local/go?" "Y" || { warn "Skip install Go"; return 1; }

    local tmp; tmp="$(mktemp -d)"
    trap 'rm -rf "$tmp"' EXIT
    curl -fsSL --retry 3 -o "${tmp}/${tarball}" "$url"

    sudo_run rm -rf /usr/local/go
    sudo_run tar -C /usr/local -xzf "${tmp}/${tarball}"

    export PATH="/usr/local/go/bin:${PATH}"

    # persist PATH
    local profile_line='export PATH="/usr/local/go/bin:$PATH"'
    local profile_file="${HOME}/.profile"
    [ -f "${HOME}/.bashrc" ] && profile_file="${HOME}/.bashrc"
    if ! grep -Fqs '/usr/local/go/bin' "$profile_file" 2>/dev/null; then
        echo "$profile_line" >> "$profile_file"
        info "PATH ditambahkan ke ${profile_file} (reload shell atau 'source ${profile_file}' setelah selesai)"
    fi

    cur="$(go_version || true)"
    if [ -n "$cur" ] && version_ge "$cur" "$GO_MIN_VERSION"; then
        ok "Go ${cur} terpasang di /usr/local/go"
    else
        err "Instalasi Go gagal — versi terdeteksi: '${cur:-none}'"
        return 1
    fi
}

# ---------- step: .env ----------
configure_env() {
    section "3. Konfigurasi .env"
    if [ "$SKIP_ENV" = 1 ]; then
        warn "Skip (--skip-env)"
        if [ ! -f "$ENV_FILE" ] && [ -f "$ENV_EXAMPLE" ]; then
            cp "$ENV_EXAMPLE" "$ENV_FILE"
            ok "Copy .env.example -> .env"
        fi
        return
    fi

    if [ -f "$ENV_FILE" ]; then
        warn ".env sudah ada di ${ENV_FILE}"
        if ! confirm "Overwrite .env yang sudah ada?" "N"; then
            ok "Pakai .env yang sudah ada"
            return
        fi
        cp "$ENV_FILE" "${ENV_FILE}.bak.$(date +%Y%m%d%H%M%S)"
        info "Backup .env lama -> ${ENV_FILE}.bak.*"
    fi

    local owners db log_level self_mode mustika_key ai_key
    owners="$(ask "Nomor owner WhatsApp (format internasional tanpa +, pisah koma utk multi)" "")"
    while [ -z "$owners" ]; do
        warn "Owner wajib diisi"
        owners="$(ask "Nomor owner WhatsApp (format internasional tanpa +)" "")"
    done
    db="$(ask "Path database SQLite" "gowa-bot.db")"
    log_level="$(ask "Log level (debug/info/warn/error)" "info")"
    self_mode="$(ask "Self mode? (true/false)" "false")"
    mustika_key="$(ask "MustikaPay API key (opsional, kosongin kalau tidak pakai)" "")"
    ai_key="$(ask "AI (Claude) API key (opsional, kosongin kalau tidak pakai)" "")"

    cat > "$ENV_FILE" <<EOF
export GOWA_BOT_OWNERS="${owners}"
export GOWA_BOT_DB="${db}"
export GOWA_BOT_LOG_LEVEL="${log_level}"
export GOWA_BOT_SELF_MODE="${self_mode}"
export GOWA_BOT_MUSTIKA_API_KEY="${mustika_key}"
export GOWA_BOT_AI_API_KEY="${ai_key}"
EOF
    chmod 600 "$ENV_FILE"
    ok ".env ditulis ke ${ENV_FILE} (mode 600)"
}

# ---------- step: build ----------
build_binary() {
    section "4. Build binary"
    if [ "$SKIP_BUILD" = 1 ]; then warn "Skip (--skip-build)"; return; fi

    cd "$REPO_DIR"
    info "go mod download"
    go mod download
    info "go build -o ${BIN_NAME}"
    go build -o "$BIN_NAME"
    ok "Binary tersedia: ${REPO_DIR}/${BIN_NAME}"
}

# ---------- step: systemd ----------
install_service() {
    section "5. Systemd service (opsional)"
    if [ "$SKIP_SERVICE" = 1 ]; then warn "Skip (--skip-service)"; return; fi
    if [ "$OS_KIND" != "linux" ] || ! have systemctl; then
        warn "systemctl tidak ditemukan, skip"
        return
    fi
    if ! confirm "Install systemd service '${BIN_NAME}.service'?" "N"; then
        info "Skip systemd"
        return
    fi

    local phone run_user
    phone="$(ask "Nomor WhatsApp untuk pairing (untuk flag -phone)" "")"
    while [ -z "$phone" ]; do
        warn "Nomor wajib diisi (untuk -phone)"
        phone="$(ask "Nomor WhatsApp untuk pairing" "")"
    done
    run_user="$(ask "Jalankan sebagai user" "${SUDO_USER:-$USER}")"

    local tmp_unit; tmp_unit="$(mktemp)"
    cat > "$tmp_unit" <<EOF
[Unit]
Description=Gowa-Bot WhatsApp bot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${run_user}
WorkingDirectory=${REPO_DIR}
EnvironmentFile=${ENV_FILE}
ExecStart=${REPO_DIR}/${BIN_NAME} -phone ${phone}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    sudo_run install -m 644 "$tmp_unit" "$SERVICE_FILE"
    rm -f "$tmp_unit"
    sudo_run systemctl daemon-reload
    ok "Service ditulis: ${SERVICE_FILE}"

    if confirm "Enable & start service sekarang?" "Y"; then
        sudo_run systemctl enable --now "${BIN_NAME}.service"
        ok "Service aktif. Cek: 'sudo systemctl status ${BIN_NAME}'"
        info "Pairing code di-emit ke journal. Cek: 'sudo journalctl -u ${BIN_NAME} -f'"
    else
        info "Untuk enable nanti: sudo systemctl enable --now ${BIN_NAME}"
    fi
}

# ---------- summary ----------
print_summary() {
    section "Selesai"
    cat <<EOF
${C_GREEN}Gowa-Bot siap dipakai.${C_RESET}

Langkah berikutnya:
  ${C_DIM}# load env${C_RESET}
  source .env

  ${C_DIM}# jalankan dengan pairing${C_RESET}
  ./${BIN_NAME} -phone <nomor_anda>

  ${C_DIM}# dashboard web (default)${C_RESET}
  ./${BIN_NAME} -phone <nomor> -web-addr :8080
  ${C_DIM}# buka http://localhost:8080${C_RESET}

Doc lengkap: README.md
EOF
}

# ---------- main ----------
main() {
    section "Gowa-Bot installer"
    detect_os
    info "OS: ${OS_KIND}  Arch: ${ARCH}  Pkg: ${PKG:-<none>}"

    if [ ! -f "$ENV_EXAMPLE" ]; then
        err "${ENV_EXAMPLE} tidak ditemukan. Jalankan installer dari root repo."
        exit 1
    fi

    install_system_deps
    install_go
    configure_env
    build_binary
    install_service
    print_summary
}

main "$@"
