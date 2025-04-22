package healthchecker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"

	resp "lyrics-library/internal/lib/api/response"
	"lyrics-library/internal/lib/logger/sl"
)

type StorageHealthChecker interface {
	Ping(ctx context.Context) error
}

func New(
	log *slog.Logger,
	pgClient StorageHealthChecker,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := pgClient.Ping(ctx); err != nil {
				log.Error("postgres health check failed", sl.Err(err))

				w.WriteHeader(http.StatusInternalServerError)

				render.JSON(w, r, resp.Error("internal error"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
