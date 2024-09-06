package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/neet-007/rss/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("CONNECTION_STRING")

	db, err := sql.Open("postgres", dbURL)

	dbQueries := database.New(db)

	config := apiConfig{
		DB: dbQueries,
	}

	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", handleReadiness)
	mux.HandleFunc("GET /v1/error", handleError)
	mux.HandleFunc("POST /v1/users", config.handleCreateUser)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server is running on host 127.0.0.1 and port %s\n", port)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server is running on host 127.0.0.1 and port %s\n", port)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "couldn't read request")
		fmt.Println("couldnt not json")
		return
	}

	params := struct {
		Name string `json:"name"`
	}{}

	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, 500, "couldn't unmarshal json")
		fmt.Printf("couldnt not json %v\n error %v\n", data, err)
		return
	}

	id := uuid.New()
	nullUUID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}
	user, err := cfg.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        nullUUID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithError(w, 500, "couldn't create user")
		fmt.Println("couldnt create user")
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
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
