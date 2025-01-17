package server

import (
	"cryptoRestTest/domain"
	"cryptoRestTest/gates/storage"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AddCurrencyHandler handles the addition of observed currencies.
//
// @Summary Add Observed Currencies
// @Description Adds a list of currencies to the observed list.
// @Tags Currencies
// @Accept json
// @Produce json
// @Param request body addCoinsReq true "Request body with coins to add"
// @Success 200 {string} string "Successfully added coins"
// @Failure 400 {string} string "Invalid input or validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /currency/add [post]
func (s *Server) AddCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.AddCurrencyHandler"

	var req addCoinsReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.log.Error(op, "Error decoding json", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	coinsStr := req.Coins
	s.log.Info("op", "connected to AddCurrencyHandler, trying to add currency id: ", coinsStr)
	coins := strings.Split(coinsStr, ",")
	err = s.coinSrv.AddObserveredCoins(coins)
	if err == domain.ErrNoVerifiedCoins { //не прошло verify coin (нет такой у coingecko)
		s.log.Debug(op, "tried to add not existing coin: ", err)
		http.Error(w, "No coin passed verification, (probably this coins don't exist?)", http.StatusBadRequest)
		return
	}
	if err == storage.ErrNoRowsAffected {
		s.log.Debug(op, "no rows affected, probably already tracked: ", err)
		http.Error(w, "Nothing happend, perhaps we already track this coin/coins", http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(op, ": error adding currency id: ", coins)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info(op, ": added coins")
	w.WriteHeader(http.StatusOK)
}

// getList returns the list of currently observed currencies.
//
// @Summary Get Observed Currencies
// @Description Retrieves a list of all observed currencies.
// @Tags Currencies
// @Produce json
// @Success 200 {object} []string "List of observed currencies"
// @Failure 500 {string} string "Internal server error"
// @Router /currency/watchlist [get]
func (s *Server) getList(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.getList"
	s.log.Info("op", "connected to getList")

	coins, err := s.coinSrv.GetObserveredCoinsList()
	if err != nil {
		s.log.Error(op, ": error getting observered coins: ", err)
		http.Error(w, "Error getting observered coins", http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(coins)
	if err != nil {
		s.log.Error(op, ": error marshaling response: ", err)
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	s.log.Info(op, "retrieved coins: ", coins)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
	return
}

// CurrencyPriceHandler retrieves the price of a currency at a specific time.
//
// @Summary Get Currency Price at Specific Time
// @Description Retrieves the price of a specific currency at a given timestamp.
// @Tags Currencies
// @Accept json
// @Produce json
// @Param coin query string true "Currency symbol (e.g., BTC)"
// @Param timestamp query string true "Timestamp in Unix format"
// @Success 200 {object} coinPriceTimeResponse "Price and timestamp of the requested currency"
// @Failure 400 {string} string "Invalid input or validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /currency/price [get]
func (s *Server) CurrencyPriceHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.CurrencyPriceHandler"

	// Извлекаем параметры из строки запроса
	coin := r.URL.Query().Get("coin")
	timestampStr := r.URL.Query().Get("timestamp")

	if coin == "" || timestampStr == "" {
		s.log.Error(op, "Missing required query parameters")
		http.Error(w, "Missing required query parameters", http.StatusBadRequest)
		return
	}

	// Парсим timestamp
	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		s.log.Error(op, "Invalid timestamp format", err)
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return
	}
	timestamp := time.Unix(timestampInt, 0).UTC()

	// Получаем цену
	price, time, err := s.coinSrv.GetTimePrice(coin, timestamp)
	if err == sql.ErrNoRows {
		s.log.Error(op, ": error getting time price: ", err)
		http.Error(w, "No price found for this coin, perhaps we don't track this coin or it doesn't exist?", http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(op, "Failed to get time price", err)
		http.Error(w, "Failed to get time price", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := coinPriceTimeResponse{
		Coin:      coin,
		Timestamp: time,
		Price:     price,
	}

	response, err := json.Marshal(resp)
	if err != nil {
		s.log.Error(op, "Failed to marshal response", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	s.log.Info(op, "Retrieved time price:", price, "for timestamp:", time)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// DeleteCurrencyHandler handles the deletion of observed currencies.
//
// @Summary Delete Observed Currencies
// @Description Deletes a list of currencies from the observed list.
// @Tags Currencies
// @Accept json
// @Produce json
// @Param request body deleteCoinsReq true "Request body with coins to delete"
// @Success 200 {string} string "Successfully deleted coins"
// @Failure 400 {string} string "Invalid input or validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /currency/remove [delete]
func (s *Server) DeleteCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.DeleteCurrencyHandler"

	var req deleteCoinsReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.log.Error(op, "Error decoding json", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.log.Info(op, "connected to DeleteCurrencyHandler, trying to delete currency id: ", req.Coin)

	coins := strings.Split(req.Coin, ",")
	if len(coins) == 0 {
		s.log.Error(op, "no coins to delete")
		http.Error(w, "No coins to delete", http.StatusBadRequest)
		return
	}
	err = s.coinSrv.DeleteObserveredCoins(coins)
	if err == storage.ErrNoRowsAffected {
		s.log.Debug(op, "no rows affected, probably wasn't it storage: ", err)
		http.Error(w, "Nothing happend, perhaps it wasn't in our tracking list?", http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(op, ": error deleting currency id: ", coins)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info(op, "deleted currency: ", coins)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted coins"))
}
