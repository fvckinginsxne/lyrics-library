package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "lyrics-library/docs"
	authGRPC "lyrics-library/internal/client/grpc/auth"
	"lyrics-library/internal/client/http/track/lyricsovh"
	"lyrics-library/internal/client/http/track/yandex"
	"lyrics-library/internal/config"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/lib/logger/slogpretty"
	authService "lyrics-library/internal/service/auth"
	"lyrics-library/internal/service/track"
	"lyrics-library/internal/storage/postgres"
	"lyrics-library/internal/storage/redis"
	"lyrics-library/internal/transport/handler/auth/login"
	"lyrics-library/internal/transport/handler/auth/register"
	"lyrics-library/internal/transport/handler/track/create"
	del "lyrics-library/internal/transport/handler/track/delete"
	"lyrics-library/internal/transport/handler/track/read"
	mwAuth "lyrics-library/internal/transport/middleware/auth"
	healthChecker "lyrics-library/internal/transport/middleware/health-checker"
	mwLogger "lyrics-library/internal/transport/middleware/logger"
)

const (
	envLocal = "local"
	envProd  = "prod"

	shutdownTimeout = 15 * time.Second
)

// @title Lyrics Library API
// @version 1.0
// @description API for getting song lyrics with translation by artist and title
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer cancel()

	dbURL := connURL(cfg)

	log.Debug("connecting to database", slog.String("url", dbURL))

	storage, err := postgres.New(dbURL)
	if err != nil {
		panic(err)
	}

	redisHost := redisHost(cfg)

	log.Debug("connecting to redis", slog.String("host", redisHost))

	cache, err := redis.New(redisHost, cfg.Redis.Password)
	if err != nil {
		panic(err)
	}

	lyricsClient := lyricsovh.New(log, cfg.LyricsAPI.URL)
	translateClient := yandex.New(log,
		cfg.TranslatorAPI.Key,
		cfg.TranslatorAPI.URL,
		cfg.TranslatorAPI.TargetLang,
	)
	authClient, err := authGRPC.New(log, cfg)
	if err != nil {
		panic(err)
	}

	trackService := track.New(
		log,
		lyricsClient,
		translateClient,
		storage,
		cache,
	)
	auth := authService.New(log, authClient)

	g := gin.New()

	g.Use(gin.Recovery())
	g.Use(healthChecker.New(log, storage))
	g.Use(mwLogger.New(log))
	g.Use(mwAuth.New(log, authClient))

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := g.Group("/auth")
	{
		authGroup.POST("/register", register.New(ctx, log, auth))
		authGroup.POST("/login", login.New(ctx, log, auth))
	}

	lyricsGroup := g.Group("/lyrics")
	{
		lyricsGroup.POST("/", create.New(ctx, log, trackService))
		lyricsGroup.GET("/", read.New(ctx, log, trackService, trackService))
		lyricsGroup.DELETE("/:uuid", del.New(ctx, log, trackService))
	}

	srv := &http.Server{
		Addr:         serverAddress(cfg),
		Handler:      g,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting server", slog.String("address", serverAddress(cfg)))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		log.Error("server error", sl.Err(err))
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown server", sl.Err(err))
	}

	if err := storage.Close(shutdownCtx); err != nil {
		log.Error("failed to close storage", sl.Err(err))
	}

	if err := cache.Close(shutdownCtx); err != nil {
		log.Error("failed to close redis", sl.Err(err))
	}

	log.Info("service stopped gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettyLogger()
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func serverAddress(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
}

func connURL(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DockerPort, cfg.DB.Name)
}

func redisHost(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.DockerPort)
}
