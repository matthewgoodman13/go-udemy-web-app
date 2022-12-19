package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"myapp/internal/driver"
	"myapp/internal/models"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"
const cssVersion = "1"

type config struct {
	port int
	env  string
	api  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
	DB            models.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting HTTP server in %s mode on port %d", app.config.env, app.config.port)
	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// Read command line flags
	flag.IntVar(&cfg.port, "port", 4000, "Server listening port")
	flag.StringVar(&cfg.env, "env", "development", "Application environment (development|production)")
	flag.StringVar(&cfg.db.dsn, "dsn", "matthewgoodman13:matthew@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN for database connection")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "URL to api")

	flag.Parse()

	// Retrieve stripe key and secret from environment variables
	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	// Set up logging
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Connect to database
	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	// Close connection when main() exits
	defer conn.Close()

	// Create map for template cache
	tc := make(map[string]*template.Template)

	// Initialize a new instance of application containing the config struct
	app := &application{
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB:            models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
