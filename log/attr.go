// Package log holds custom logging types and additional slog attributes.
package log

import (
	"io"
	"log/slog"
)

// Error returns an Attr for a error value.
func Error(value error) slog.Attr {
	return slog.Attr{Key: "err", Value: slog.StringValue(value.Error())}
}

// Noop returns a no-op logger.
func Noop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Strings returns an Attr for a slice of strings.
func Strings(key string, values []string) slog.Attr {
	return slog.Attr{Key: key, Value: slog.AnyValue(values)}
}
