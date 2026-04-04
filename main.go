package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/siddhu5pute/chirpy/internal/database"
)

// apiConfig holds DB queries, platform, and JWT secret
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

// ─────────────────────────────────────────────
// Response Structs
// ─────────────────────────────────────────────

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// ─────────────────────────────────────────────
// Middleware
// ─────────────────────────────────────────────

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// ─────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DB_URL")
	dbPlatform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	dbQueries := database.New(db)

	apiConf := apiConfig{
		dbQueries: dbQueries,
		platform:  dbPlatform,
		jwtSecret: jwtSecret,
		polkaKey:  polkaKey,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiConf.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))

	mux.HandleFunc("GET /api/healthz", handlerOne)
	mux.HandleFunc("GET /admin/metrics", apiConf.handlerReqCount)
	mux.HandleFunc("POST /admin/reset", apiConf.handlerResetCount)
	mux.HandleFunc("POST /api/users", apiConf.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiConf.handlerUserLogin)
	mux.HandleFunc("POST /api/chirps", apiConf.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiConf.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConf.handlerGetChirp)
	mux.HandleFunc("POST /api/refresh", apiConf.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiConf.handlerRevoke)
	mux.HandleFunc("PUT /api/users", apiConf.handlerUpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConf.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiConf.handlerUpgradeUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
