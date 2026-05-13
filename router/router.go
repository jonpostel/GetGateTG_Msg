package router

import (
	"getGate/handlers"
	"net/http"
)

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func SetupRoutes() {
	http.HandleFunc("/api/config", CORS(handlers.ConfigHandler))
	http.HandleFunc("/api/fear-greed", CORS(handlers.FearGreedHandler))
	http.HandleFunc("/api/market", CORS(handlers.GetMarketData))
}
