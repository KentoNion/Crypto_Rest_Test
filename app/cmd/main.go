package main

import (
	"context"
	"cryptoRestTest/domain"
	coingecko "cryptoRestTest/gates/providers"
	"cryptoRestTest/gates/server"
	"cryptoRestTest/gates/storage"
	"cryptoRestTest/internal/config"
	"cryptoRestTest/internal/logger"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //драйвер postgres
	goose "github.com/pressly/goose/v3"
	"net/http"
	"os"
	"time"
)

func main() {
	//инициализация конфига
	cfg := config.MustLoad()

	//инициализация логгера
	log := logger.MustInitLogger(cfg)
	log.Debug("logger started in debug mode")

	//инициализация бд
	dbhost := os.Getenv("DB_HOST") //DB_HOST прописывается в docker_compose, если его там нет, значит считается из конфига
	if dbhost == "" {
		dbhost = cfg.DB.Host
	}
	connStr := fmt.Sprintf("user=%s password=%s dbname=coins host=%s sslmode=%s timezone=UTC", cfg.DB.User, cfg.DB.Pass, dbhost, cfg.DB.Ssl)
	conn, err := sqlx.Connect("postgres", connStr) //подключение к бд
	if err != nil {
		panic(err)
	}
	store := storage.NewDB(conn, log)

	//накатка миграций
	migrationsPath := os.Getenv("MIGRATIONS_PATH") //для докера
	if migrationsPath == "" {
		migrationsPath = "./gates\\storage\\migrations"
	}
	//err = goose.Down(conn.DB, migrationsPath)
	err = goose.Up(conn.DB, migrationsPath)
	if err != nil {
		panic(err)
	}

	//инициализация provider client
	provider := coingecko.NewClient(context.Background(), cfg, log)

	//инициализация watcher
	watcher := domain.NewWatcher(context.Background(), store, log, provider, cfg)

	//запуск горутины по отслеживанию монет
	go func(watcher *domain.Watcher) {
		observeTicker := time.NewTicker(cfg.CoinsWatcher.Cooldown)
		for {
			select {
			case <-observeTicker.C:
				err = watcher.ScanPrices()
				if err != nil {
					log.Warn("------------------WARNING, ScanPrices failed!--------------------------")
				}
			}
		}
	}(watcher)

	//настройка и запуск REST сервера
	router := chi.NewRouter()
	_ = server.NewServer(router, store, log, cfg, watcher)
	restServerAddr := cfg.Rest.Host + ":" + cfg.Rest.Port //получение адреса rest сервера из конфига
	err = http.ListenAndServe(restServerAddr, router)
	if err != nil {
		panic(err)
	}
}
