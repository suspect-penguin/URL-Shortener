package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/komoru/url-shortener/internal/cache"
	"github.com/komoru/url-shortener/internal/config"
	"github.com/komoru/url-shortener/internal/handler"
	"github.com/komoru/url-shortener/internal/repository"
	"github.com/komoru/url-shortener/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("подключение к БД: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping БД: %v", err)
	}
	log.Println("подключение к БД установлено")

	c := cache.New(5 * time.Minute)
	defer c.Stop()

	repo := repository.NewPostgres(pool)
	svc := service.New(repo, c, cfg.BaseURL)
	h := handler.New(svc)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      h.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("сервер запущен на %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка сервера: %v", err)
		}
	}()

	<-quit
	log.Println("завершение работы...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("принудительное завершение: %v", err)
	}
	log.Println("сервер остановлен")
}
