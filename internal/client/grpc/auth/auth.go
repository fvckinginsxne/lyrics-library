package auth

import (
	"context"
	"fmt"
	"log/slog"

	ssov1 "github.com/fvckinginsxne/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"lyrics-library/internal/client"
	"lyrics-library/internal/config"
)

type Client struct {
	api ssov1.AuthClient
	log *slog.Logger
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) (*Client, error) {
	const op = "client.grpc.auth.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded, codes.NotFound),
		grpcretry.WithMax(uint(cfg.Auth.Retries)),
		grpcretry.WithPerRetryTimeout(client.RequestTimeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	addr := clientAddress(cfg)

	log.Debug("Auth service address:", slog.String("address", addr))

	cc, err := grpc.NewClient(clientAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(*log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: ssov1.NewAuthClient(cc),
	}, nil
}

func InterceptorLogger(l slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func clientAddress(cfg *config.Config) string {
	return cfg.Auth.Host + ":" + cfg.Auth.Port
}
