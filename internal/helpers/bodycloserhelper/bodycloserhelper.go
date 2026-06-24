package bodycloserhelper

import (
	"log/slog"
)

func BodyCloseResponse(err error) {
	if err != nil {
		slog.Error("failed to close response body", "error", err)
	}
}
