#!/bin/sh
set -eu

# Resolve script directory in a POSIX-compatible way.
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)

# Исправленные пути в соответствии с вашей структурой проекта
MAIN_DIR="$SCRIPT_DIR/cmd/send-data"
OUTPUT_BIN="$SCRIPT_DIR/send-data"

if ! command -v go >/dev/null 2>&1; then
    echo "error: go toolchain not found in PATH" >&2
    exit 1
fi

# Проверяем наличие директории вместо конкретного файла
if [ ! -d "$MAIN_DIR" ]; then
    echo "error: source directory not found: $MAIN_DIR" >&2
    exit 1
fi

# Оптимизация сборки:
# -ldflags="-w -s" убирает отладочную информацию и уменьшает размер бинарника.
# Флаг -v можно убрать, если не нужен подробный вывод компиляции.
go build -ldflags="-w -s" -o "$OUTPUT_BIN" "$MAIN_DIR"

echo "built: $OUTPUT_BIN"

