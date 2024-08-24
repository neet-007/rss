package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", handleReadiness)
	mux.HandleFunc("GET /v1/error", handleError)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server is running on host 127.0.0.1 and port %s\n", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
	log.Printf("server is running on host 127.0.0.1 and port %s\n", port)
}

func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	respondWithJSON(w, 200, struct {
		Status string
	}{Status: "ok"})
}

func handleError(w http.ResponseWriter, _ *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	jsonResponse, err := json.Marshal(payload)
	if err != nil {
		log.Println("An error occurred when converting to JSON:", err)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("An error occurred when converting to JSON"))
		return
	}
	println(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonResponse)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	log.Println("error:", msg)
	respondWithJSON(w, code, struct {
		Error string `json:"error"`
	}{Error: msg})
}
