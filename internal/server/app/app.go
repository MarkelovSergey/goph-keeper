// Package app выполняет инициализацию и запуск сервера.
package app

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

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/MarkelovSergey/goph-keeper/internal/server/config"
	"github.com/MarkelovSergey/goph-keeper/internal/server/handler"
	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository/postgres"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

const (
	tokenTTL     = 24 * time.Hour
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

// App инкапсулирует HTTP-сервер и соединение с базой данных.
type App struct {
	cfg    *config.Config
	server *http.Server
	db     *pgxpool.Pool
}

// New создаёт и конфигурирует приложение.
func New(cfg *config.Config) (*App, error) {
	// Подключение к PostgreSQL
	db, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("БД недоступна: %w", err)
	}
	slog.Info("Подключение к базе данных установлено")

	// Репозитории
	userRepo := postgres.NewUserRepository(db)
	credRepo := postgres.NewCredentialRepository(db)

	// Сервисы
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, tokenTTL)
	credSvc := service.NewCredentialService(credRepo)

	// Обработчики
	authHandler := handler.NewAuthHandler(authSvc)
	credHandler := handler.NewCredentialHandler(credSvc)

	// Маршрутизатор
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc))
		r.Get("/api/credentials", credHandler.List)
		r.Post("/api/credentials", credHandler.Create)
		r.Get("/api/credentials/{id}", credHandler.Get)
		r.Put("/api/credentials/{id}", credHandler.Update)
		r.Delete("/api/credentials/{id}", credHandler.Delete)
	})

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return &App{
		cfg,
		srv,
		db,
	}, nil
}

// Start запускает сервер, ожидает сигнала ОС и выполняет graceful shutdown.
func (a *App) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serveErr := make(chan error, 1)
	go func() {
		if a.cfg.TLSCertPath != "" && a.cfg.TLSKeyPath != "" {
			slog.Info("Запуск HTTPS-сервера", "addr", a.cfg.ListenAddr)
			serveErr <- a.server.ListenAndServeTLS(a.cfg.TLSCertPath, a.cfg.TLSKeyPath)
		} else {
			slog.Info("Запуск HTTP-сервера", "addr", a.cfg.ListenAddr)
			serveErr <- a.server.ListenAndServe()
		}
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("ошибка сервера: %w", err)
		}
	case <-quit:
		slog.Info("Остановка сервера...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer a.db.Close()
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка graceful shutdown: %w", err)
	}
	slog.Info("Сервер остановлен")
	return nil
}
