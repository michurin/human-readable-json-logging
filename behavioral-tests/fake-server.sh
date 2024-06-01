#!/bin/sh

echo '{"time":"2024-12-02T12:00:00-05:00","level":"INFO","source":{"function":"main.main","file":"/Users/slog/main.go","line":14},"msg":"OK","user":1}'
echo 'Broken raw error message'
echo '{"time":"2024-12-02T12:05:00-05:00","level":"INFO","source":{"function":"main.main","file":"/Users/slog/main.go","line":14},"msg":"OK","user":2}'
