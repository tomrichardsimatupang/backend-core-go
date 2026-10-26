package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

type ctxKey struct{}

// New membuat logger slog. format "text" untuk development, selain itu JSON.
func New(level, format string, w io.Writer) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lvl}
	if strings.EqualFold(format, "text") {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

// WithContext menyimpan logger (biasanya sudah berisi request_id) ke context.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext mengambil logger dari context, fallback ke slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
