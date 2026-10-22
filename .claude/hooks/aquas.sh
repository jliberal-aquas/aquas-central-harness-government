#!/bin/sh
# Lanza aquas para los hooks de Claude Code.
# - Temporales de Go dentro de ./.tmp (regla de oro: nada se escribe fuera del repo).
# - Ejecuta el binario directo: `go run` convierte el exit 2 (bloquear) en exit 1 (no bloquea).
# - Binario por proceso: hooks en paralelo no pisan el mismo .exe (Windows no reemplaza uno en uso).
# - Falla cerrado (E000/F006/S001): si aquas no compila, corre la última versión que compiló
#   (.tmp/aquas-ultimo.exe) para que guardia, memoria y cierre sigan vigilando y se pueda
#   arreglar el código. Sin versión anterior, sale 2 y bloquea.
cd "${CLAUDE_PROJECT_DIR:-.}" || exit 2
mkdir -p .tmp
export GOTMPDIR="$PWD/.tmp" TMP="$PWD/.tmp" TEMP="$PWD/.tmp"
bin=".tmp/aquas-hook-$$.exe"
ultimo=".tmp/aquas-ultimo.exe"
if go build -o "$bin" ./cmd/aquas; then
	cp -f "$bin" "$ultimo" 2>/dev/null || true
elif [ -f "$ultimo" ]; then
	echo "aquas: el código no compila; el hook usa la última versión que compiló" >&2
	cp -f "$ultimo" "$bin" || exit 2
else
	echo "aquas: el código no compila y no hay versión anterior; el hook bloquea (falla cerrado)" >&2
	exit 2
fi
"./$bin" "$@"
codigo=$?
rm -f "$bin"
exit $codigo
