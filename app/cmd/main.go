package main

import (
	"cryptoRestTest/internal/config"
	"cryptoRestTest/internal/logger"
)

func main() {
	//инициализация конфига
	cfg := config.MustLoad()

	//инициализация логгера
	log := logger.MustInitLogger(cfg)
	log.Debug("logger started in debug mode")

	//инициализация бд

	//инициализация coingeko

	//запуск горутины по отслеживанию монет

	//запуск REST сервера
}
