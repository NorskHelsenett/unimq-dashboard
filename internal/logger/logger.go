package logger

import (
	"context"
	"log/slog"
	"os"
	"path"
)

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, record slog.Record) error {

	if requestID, ok := ctx.Value("request_id").(string); ok {
		record.Add("request_id", requestID)
	}

	return h.Handler.Handle(ctx, record)
}

func SetupLogger() {

	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				s := a.Value.Any().(*slog.Source)
				s.File = path.Base(s.File)
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stdout, handlerOpts)
	logger := slog.New(&ContextHandler{Handler: handler})

	slog.SetDefault(logger)
}
