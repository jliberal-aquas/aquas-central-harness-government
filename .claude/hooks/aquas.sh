#!/bin/sh
# Lanza aquas para los hooks de Claude Code.
# - Temporales de Go dentro de ./.tmp (regla de oro: nada se escribe fuera del repo).
# - Ejecuta el binario directo: `go run` convierte el exit 2 (bloquear) en exit 1 (no bloquea).
# - Binario por proceso: hooks en paralelo no pisan el mismo .exe (Windows no reemplaza uno en uso).
# Si no compila, sale 1: Claude Code lo muestra pero no bloquea, para no dejar la sesión sin poder editar.
cd "${CLAUDE_PROJECT_DIR:-.}" || exit 1
mkdir -p .tmp
export GOTMPDIR="$PWD/.tmp" TMP="$PWD/.tmp" TEMP="$PWD/.tmp"
bin=".tmp/aquas-hook-$$.exe"
go build -o "$bin" ./cmd/aquas || exit 1
"./$bin" "$@"
codigo=$?
rm -f "$bin"
exit $codigo
