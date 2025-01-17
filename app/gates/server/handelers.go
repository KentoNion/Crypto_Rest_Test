package server

import (
	"cryptoRestTest/domain"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
	if err != nil {
		s.log.Error(op, ": error adding currency id: ", coins)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info(op, ": added coins")
	w.WriteHeader(http.StatusOK)
}

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

func (s Server) CurrencyPriceHandler(w http.ResponseWriter, r *http.Request) {
	const op = "gates.Server.CurrencyPriceHandler"

	var coinTime coinPriceTimeRequest
	err := json.NewDecoder(r.Body).Decode(&coinTime)
	s.log.Debug(op, "decoded body: ", coinTime)
	if err != nil {
		s.log.Error(op, "Error decoding json", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = coinTime.Validate()
	if err != nil {
		s.log.Debug(op, "Error validating coinTime", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	coin := coinTime.Coin
	timestampInt, err := strconv.ParseInt(coinTime.Timestamp, 10, 64)
	if err != nil {
		s.log.Error(op, "failed to parse timestamp", err)
		http.Error(w, "failed to parse time", http.StatusBadRequest)
		return
	}
	timestamp := time.Unix(timestampInt, 0).UTC()
	price, time, err := s.coinSrv.GetTimePrice(coin, timestamp)
	if err != nil {
		s.log.Error(op, "failed to get time price", err)
		http.Error(w, "failed to get time price", http.StatusBadRequest)
		return
	}
	var resp coinPriceTimeResponse
	resp.Timestamp, resp.Price = time, price
	response, err := json.Marshal(resp)
	if err != nil {
		s.log.Error(op, "failed to marshal response: ", err)
		http.Error(w, "failed to marshal response", http.StatusBadRequest)
		return
	}
	s.log.Info(op, "retrieved time price: ", price, "for timestamp: ", time)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

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
	if err != nil {
		s.log.Error(op, ": error deleting currency id: ", coins)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info(op, "deleted currency: ", coins)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted coins"))
}
