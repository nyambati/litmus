#!/bin/sh
set -e

# Litmus Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/nyambati/litmus/main/scripts/install.sh | sh
# Or to specify a directory: curl -sSL https://raw.githubusercontent.com/nyambati/litmus/main/scripts/install.sh | BINDIR=$HOME/bin sh

OWNER="nyambati"
REPO="litmus"
BINARY="litmus"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)   OS='linux';;
  darwin*)  OS='darwin';;
  msys*)    OS='windows'; BINARY="${BINARY}.exe";;
  *)        echo "OS ${OS} not supported"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64*)  ARCH='amd64';;
  arm64*|aarch64*) ARCH='arm64';;
  *)        echo "Architecture ${ARCH} not supported"; exit 1;;
esac

# Set installation directory
if [ -z "$BINDIR" ]; then
    # Default to ~/.local/bin which is standard for user-local binaries
    BINDIR="$HOME/.local/bin"
fi

# Get latest version from GitHub API
echo "Finding latest version of ${REPO}..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/${OWNER}/${REPO}/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not find latest release"
    exit 1
fi

echo "Latest version is ${LATEST_RELEASE}"

# Construct download URL
FILENAME="${REPO}_${LATEST_RELEASE#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${OWNER}/${REPO}/releases/download/${LATEST_RELEASE}/${FILENAME}"

# Create temp directory
TMP_DIR=$(mktemp -d)
cd "${TMP_DIR}"

echo "Downloading ${URL}..."
curl -sSL -O "${URL}"

echo "Extracting..."
tar -xzf "${FILENAME}"

echo "Installing to ${BINDIR}..."
mkdir -p "${BINDIR}"
mv "${BINARY}" "${BINDIR}/"

echo "Successfully installed ${BINARY} ${LATEST_RELEASE} to ${BINDIR}"

# Check if BINDIR is in PATH
if ! echo "$PATH" | grep -q "${BINDIR}"; then
    echo ""
    echo "WARNING: ${BINDIR} is not in your PATH."
    echo "You may need to add it to your shell configuration (.bashrc, .zshrc, etc.):"
    echo "  export PATH=\$PATH:${BINDIR}"
fi

"${BINDIR}/${BINARY}" --version || "${BINDIR}/${BINARY}" --help | head -n 1
