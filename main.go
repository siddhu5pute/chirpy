package main

// Postgres driver (needed for database/sql to talk to Postgres)
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/siddhu5pute/chirpy/internal/database"

	"time"

	"github.com/google/uuid"
)

// Load .env file

// IMPORT YOUR SQLC GENERATED PACKAGE

// apiConfig now ALSO holds DB queries
type apiConfig struct {
	fileserverHits atomic.Int32

	// NEW: SQLC queries object
	dbQueries *database.Queries
	platform  string
}

// Middleware to count hits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Show metrics in browser
func (cfg *apiConfig) handlerReqCount(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>`, cfg.fileserverHits.Load())))
}

// Reset counter
func (cfg *apiConfig) handlerResetCount(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
		return
	}
	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

//Create User

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	r_email := struct {
		Email string `json:"email"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&r_email)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), r_email.Email)
	if err != nil {
		respondWithError(w, 500, "Couldn't create user")
		return
	}

	respondWithJSON(w, 201, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

// Create CHIRP
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	c_body := struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&c_body)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters")
		return
	}
	validatedB, err := validateChirp(c_body.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validatedB,
		UserID: c_body.UserID,
	})
	if err != nil {
		respondWithError(w, 500, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

//Get all chirps

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	var structuredchirps []Chirp
	for _, chirp := range chirps {
		strChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		structuredchirps = append(structuredchirps, strChirp)
	}
	respondWithJSON(w, 200, structuredchirps)
}

// Get chirp
func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func main() {

	// ✅ STEP 1: Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// ✅ STEP 2: Get DB URL from environment
	dbURL := os.Getenv("DB_URL")
	dbPLATFORM := os.Getenv("PLATFORM")

	// ✅ STEP 3: Open DB connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	// ✅ STEP 4: Create SQLC Queries instance
	dbQueries := database.New(db)

	// ✅ STEP 5: Inject into apiConfig
	apiConf := apiConfig{
		dbQueries: dbQueries,
		platform:  dbPLATFORM,
	}

	// ROUTES
	mux := http.NewServeMux()

	mux.Handle("/app/", apiConf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerOne)
	mux.HandleFunc("GET /admin/metrics", apiConf.handlerReqCount)
	mux.HandleFunc("POST /admin/reset", apiConf.handlerResetCount)
	mux.HandleFunc("POST /api/chirps", apiConf.handlerCreateChirp)
	mux.HandleFunc("POST /api/users", apiConf.handlerCreateUser)
	mux.HandleFunc("GET /api/chirps", apiConf.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConf.handlerGetChirp)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
