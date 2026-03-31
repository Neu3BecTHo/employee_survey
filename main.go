package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	Router  *mux.Router
	DB      *sql.DB
	DevMode bool
}

func (a *App) Initialize(user, password, dbname, host string, port int) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Check DEV_MODE environment variable
	devMode := os.Getenv("DEV_MODE")
	a.DevMode = strings.ToLower(devMode) == "true" || devMode == "1"
	if a.DevMode {
		log.Println("Running in DEVELOPMENT mode (unminified files)")
	} else {
		log.Println("Running in PRODUCTION mode (minified files)")
	}

	connectionString := "host=" + host + " port=" + strconv.Itoa(port) + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	// Force IPv4 connection
	if host == "localhost" || host == "127.0.0.1" {
		connectionString += " hostaddr=127.0.0.1"
	}

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

// Static file server with proper MIME type handling
func (a *App) staticFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Static file request: %s", path)

	// Remove /static/ prefix and construct file path
	filePath := "." + path

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		http.NotFound(w, r)
		return
	}

	// Get file extension and set content type explicitly
	ext := filepath.Ext(filePath)
	var contentType string
	switch ext {
	case ".js":
		contentType = "application/javascript"
		log.Printf("Setting MIME type for JS file: %s -> %s", filePath, contentType)
	case ".css":
		contentType = "text/css"
		log.Printf("Setting MIME type for CSS file: %s -> %s", filePath, contentType)
	case ".html":
		contentType = "text/html"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	default:
		contentType = "application/octet-stream"
	}

	// Set proper MIME type and write response
	w.Header().Set("Content-Type", contentType)
	log.Printf("Final Content-Type header: %s, file size: %d bytes", contentType, len(content))
	w.Write(content)
}

// customFileServer wraps http.FileServer and sets proper MIME types
type customFileServer struct {
	root http.FileSystem
}

func (cfs *customFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the file extension from URL
	ext := filepath.Ext(r.URL.Path)

	// Set Content-Type based on extension before serving
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Serve the file
	http.FileServer(cfs.root).ServeHTTP(w, r)
}

func (a *App) initializeRoutes() {
	// Register MIME types explicitly (needed for Alpine Linux in Docker)
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".svg", "image/svg+xml")

	log.Println("MIME types registered successfully")

	// Public API endpoints (JSON data - must be BEFORE frontend routes)
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/surveys", a.getSurveys).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/data", a.getSurvey).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/responses", a.submitResponse).Methods("POST")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/results/data", a.getResults).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/check-status", a.checkSurveyResponseStatus).Methods("GET")
	a.Router.HandleFunc("/surveys/my/data", a.getMyResponses).Methods("GET")
	a.Router.HandleFunc("/surveys/responses/{responseId:[0-9]+}/data", a.getResponseDetail).Methods("GET")

	// Admin API endpoints
	a.Router.Handle("/surveys", a.requireRole("admin")(http.HandlerFunc(a.createSurvey))).Methods("POST")
	a.Router.Handle("/surveys/{id:[0-9]+}", a.requireRole("admin")(http.HandlerFunc(a.updateSurvey))).Methods("PUT")
	a.Router.Handle("/surveys/{id:[0-9]+}/questions", a.requireRole("admin")(http.HandlerFunc(a.createQuestion))).Methods("POST")
	a.Router.Handle("/surveys/{id:[0-9]+}/questions/{questionId:[0-9]+}", a.requireRole("admin")(http.HandlerFunc(a.updateQuestion))).Methods("PUT")
	a.Router.Handle("/surveys/{id:[0-9]+}/questions/{questionId:[0-9]+}", a.requireRole("admin")(http.HandlerFunc(a.deleteQuestion))).Methods("DELETE")
	a.Router.Handle("/surveys/{id:[0-9]+}/open", a.requireRole("admin")(http.HandlerFunc(a.openSurvey))).Methods("POST")
	a.Router.Handle("/surveys/{id:[0-9]+}/close", a.requireRole("admin")(http.HandlerFunc(a.closeSurvey))).Methods("POST")

	// Frontend routes (AFTER API routes)
	a.Router.HandleFunc("/", a.serveIndex).Methods("GET")
	a.Router.HandleFunc("/login", a.serveLogin).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/take", a.serveSurvey).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/already-responded", a.serveAlreadyResponded).Methods("GET")
	a.Router.HandleFunc("/surveys/responses/{responseId:[0-9]+}", a.serveMyResponseDetail).Methods("GET")
	a.Router.HandleFunc("/admin/surveys", a.serveAdminSurveys).Methods("GET")
	a.Router.HandleFunc("/admin/surveys/{id:[0-9]+}", a.serveAdminSurvey).Methods("GET")
	a.Router.HandleFunc("/admin/surveys/{id:[0-9]+}/results", a.serveResults).Methods("GET")
	a.Router.HandleFunc("/surveys/{id:[0-9]+}/results", a.serveResults).Methods("GET")
	a.Router.HandleFunc("/surveys/my", a.serveMyResponses).Methods("GET")

	// Static files with custom MIME type handling
	a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", &customFileServer{root: http.Dir("./static/")}))
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

// Template data with DevMode type
type TemplateData struct {
	DevMode bool
}

// renderTemplate renders HTML template with DevMode flag
func (a *App) renderTemplate(w http.ResponseWriter, templatePath string) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		DevMode: a.DevMode,
	}

	tmpl.Execute(w, data)
}

// Frontend page handlers
func (a *App) serveIndex(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/index.html")
}

func (a *App) serveLogin(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/login.html")
}

func (a *App) serveSurvey(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/take-survey.html")
}

func (a *App) serveAlreadyResponded(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/already-responded.html")
}

func (a *App) serveAdminSurveys(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/index.html")
}

func (a *App) serveAdminSurvey(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/admin-survey.html")
}

func (a *App) serveResults(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/survey-results.html")
}

func (a *App) serveMyResponses(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/my-responses.html")
}

func (a *App) serveMyResponseDetail(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, "./static/templates/my-response-detail.html")
}
