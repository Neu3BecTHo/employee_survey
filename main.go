package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname, host string, port int) {
	connectionString := "host=" + host + " port=" + strconv.Itoa(port) + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Run migrations
	if err := a.runMigrations(); err != nil {
		log.Fatal("Migration failed:", err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) runMigrations() error {
	driver, err := postgres.WithInstance(a.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/surveys", a.getSurveys).Methods("GET")
	a.Router.HandleFunc("/surveys", a.createSurvey).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}", a.getSurvey).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}", a.updateSurvey).Methods("PUT")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/questions", a.createQuestion).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/open", a.openSurvey).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/close", a.closeSurvey).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/responses", a.submitResponse).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/results", a.getResults).Methods("GET")

	// Static files
	a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Frontend routes
	a.Router.HandleFunc("/", a.serveIndex).Methods("GET")
	a.Router.HandleFunc("/login", a.serveLogin).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/take", a.serveSurvey).Methods("GET")
	a.Router.HandleFunc("/admin/surveys", a.serveAdminSurveys).Methods("GET")
	a.Router.HandleFunc("/admin/surveys/{id:[0-9]+}", a.serveAdminSurvey).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/results", a.serveResults).Methods("GET")
}

func (a *App) Run(addr string) {
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// Middleware to check user role
func (a *App) requireRole(role string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := r.Header.Get("X-User-Id")
			if userIDStr == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			var userRole string
			err = a.DB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&userRole)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			if userRole != role && role != "any" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add user info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", userID)
			ctx = context.WithValue(ctx, "user_role", userRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	a := App{}

	// Get database config from environment
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "survey_user")
	password := getEnv("DB_PASSWORD", "survey_pass")
	dbname := getEnv("DB_NAME", "survey_app")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Invalid DB_PORT")
	}

	a.Initialize(user, password, dbname, host, port)
	a.Run(":8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper function to respond with error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Frontend page handlers
func (a *App) serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func (a *App) serveLogin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/login.html")
}

func (a *App) serveSurvey(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/take-survey.html")
}

func (a *App) serveAdminSurveys(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func (a *App) serveAdminSurvey(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html") // For now, reuse index.html
}

func (a *App) serveResults(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/survey-results.html")
}
