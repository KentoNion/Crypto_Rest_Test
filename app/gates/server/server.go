package server

import (
	"context"
	_ "cryptoRestTest/docs"
	"cryptoRestTest/domain"
	"cryptoRestTest/gates/storage"
	"cryptoRestTest/internal/config"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
)

type Server struct {
	db      *storage.Store
	ctx     context.Context
	log     *slog.Logger
	cfg     *config.Config
	coinSrv *domain.Watcher
}

func NewServer(r *chi.Mux, db *storage.Store, log *slog.Logger, conf *config.Config, watcher *domain.Watcher) *Server {
	const op = "gates.Server.NewServer"
	server := &Server{
		db:      db,
		ctx:     context.Background(),
		log:     log,
		cfg:     conf,
		coinSrv: watcher,
	}

	// Настройка маршрутов для эндпоинтов
	r.Post("/currency/add", server.AddCurrencyHandler)
	r.Delete("/currency/remove", server.DeleteCurrencyHandler)
	r.Get("/currency/price", server.CurrencyPriceHandler)
	r.Get("/currency/watchlist", server.getList)

	// Настройка Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Указываем путь к документации
	))

	server.log.Info(op, "router configured")
	return server
}
