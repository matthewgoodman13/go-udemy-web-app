package main

import (
	"flag"
	"fmt"
	"log"
	"myapp/internal/driver"
	"myapp/internal/models"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	secretkey string // to sign URLs
	frontend  string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
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

	app.infoLog.Printf("Starting Backend server in %s mode on port %d", app.config.env, app.config.port)
	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// Read command line flags
	flag.IntVar(&cfg.port, "port", 4001, "Server listening port")
	flag.StringVar(&cfg.db.dsn, "dsn", "matthewgoodman13:matthew@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN for database connection")
	flag.StringVar(&cfg.env, "env", "development", "Application environment (development|production|maintenance)")

	flag.StringVar(&cfg.smtp.host, "smtphost", "sandbox.smtp.mailtrap.io", "Host for smtp server")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "Port for smtp server")
	flag.StringVar(&cfg.smtp.username, "smtpusername", "a0e5ee79037570", "Username for smtp server")
	flag.StringVar(&cfg.smtp.password, "smtppassword", "87d6f0e74a890a", "Password for smtp server")

	flag.StringVar(&cfg.secretkey, "secret", "x6Z2c9H5F1B8g7L9A3p7D1W8k2E6h3R9", "Secret Key")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "URL to frontend")

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

	// Initialize a new instance of application containing the config struct
	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
