package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	frontend string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
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

	app.infoLog.Printf("Starting Invoice Microservice on port %d", app.config.port)
	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// Read command line flags
	flag.IntVar(&cfg.port, "port", 5000, "Server listening port")

	flag.StringVar(&cfg.smtp.host, "smtphost", "sandbox.smtp.mailtrap.io", "Host for smtp server")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "Port for smtp server")
	flag.StringVar(&cfg.smtp.username, "smtpusername", "a0e5ee79037570", "Username for smtp server")
	flag.StringVar(&cfg.smtp.password, "smtppassword", "87d6f0e74a890a", "Password for smtp server")

	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "URL to frontend")

	flag.Parse()

	// Set up logging
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize a new instance of application containing the config struct
	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}

	// Create Invoice Directory if it doesn't exist
	app.CreateDirIfNotExist("./invoices")

	// Start the HTTP server
	err := app.serve()
	if err != nil {
		errorLog.Fatal(err)
	}
}
