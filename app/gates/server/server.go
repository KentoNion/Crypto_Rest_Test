package server

import (
	"context"
	"cryptoRestTest/gates/storage"
	"cryptoRestTest/internal/config"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
)

type Server struct {
	db      *storage.Store
	context context.Context
	log     *slog.Logger
	cfg     *config.Config
}

func NewServer(r *chi.Mux, db *storage.Store, log *slog.Logger, conf *config.Config) *Server {
	const op = "gates.Server.NewServer"
	server := &Server{
		db:      db,
		context: context.Background(),
		log:     log,
		cfg:     conf,
	}

	r.Put("/currency/add", AddCurrencyHandler)
	r.Delete("/currency/remove", DeleteCurrencyHandler)
	r.Get("currency/price/{id}", CurrencyPriceHandler)

	//swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json")))

	server.log.Info(op, "router configured")
	return server
}
