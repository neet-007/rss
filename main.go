package main

import (
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
	"github.com/neet-007/rss/internal/auth"
	"github.com/neet-007/rss/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

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
	mux.HandleFunc("GET /v1/users", config.middlewareAuth(config.handleGetUserByAPI))
	mux.HandleFunc("POST /v1/feeds", config.middlewareAuth(config.handleCreateFeed))
	mux.HandleFunc("GET /v1/feeds", config.handleFetchAllFeeds)
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", config.middlewareAuth(config.handleFollowFeedDelete))
	mux.HandleFunc("GET /v1/feed_follows", config.middlewareAuth(config.handleFollowFeedGet))
	mux.HandleFunc("POST /v1/feed_follows", config.middlewareAuth(config.handleFollowFeedFollow))
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

func (cgf *apiConfig) handleFollowFeedGet(w http.ResponseWriter, r *http.Request, user database.User) {
	feed, err := cgf.DB.GetFeedFollowsForUser(r.Context(), user.ID.UUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldt find feed")
		return
	}

	respondWithJSON(w, http.StatusCreated, feed)
}

func (cfg *apiConfig) handleFollowFeedDelete(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowIDStr := r.PathValue("feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}

	err = cfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{UserID: user.ID.UUID, ID: feedFollowID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not delete feed")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}

func (cgf *apiConfig) handleFollowFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldnt read body")
		return
	}

	params := struct {
		FeedId uuid.UUID `json:"feed_id"`
	}{}

	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not unmarshal json")
		return
	}

	feed, err := cgf.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID.UUID,
		FeedID:    params.FeedId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create feed")
		return
	}

	respondWithJSON(w, http.StatusCreated, feed)
}

func (cfg *apiConfig) handleFetchAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.FetchAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not fetch feeds")
		return
	}

	respondWithJSON(w, http.StatusOK, feeds)
}

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := auth.GetAPIKey(r.Header)

		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		user, err := cfg.DB.GetUserByAPIKey(r.Context(), key)

		if err != nil {
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}

		handler(w, r, user)
	}
}

func (cgf *apiConfig) handleCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldnt read body")
		return
	}

	params := struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}{}

	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not unmarshal json")
		return
	}

	id := uuid.New()
	nullUUID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}
	feed, err := cgf.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        nullUUID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.Url,
		UserID:    user.ID.UUID,
	})

	feedFollow, err := cgf.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID.UUID,
		FeedID:    feed.ID.UUID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create feed")
		return
	}

	respondWithJSON(w, http.StatusCreated, struct {
		Feed        database.Feed       `json:"feed"`
		Feed_Follow database.FeedFollow `json:"feed_follow"`
	}{Feed: feed, Feed_Follow: feedFollow})
}

func (cfg *apiConfig) handleGetUserByAPI(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
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
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
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
