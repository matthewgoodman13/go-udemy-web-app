package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// writeJSON is a helper that writes JSON data to the client.
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {

	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)
	return nil
}

// readJSON is a helper that reads JSON data from the client.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {

	maxBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON object")
	}

	return nil
}

// badRequest is a helper that sends a Bad Request response to the client.
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
	return nil
}

// invalidCredentials is a helper that sends an Invalid Credentials response to the client.
func (app *application) invalidCredentials(w http.ResponseWriter) error {

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = "Invalid authentication credentials"

	err := app.writeJSON(w, http.StatusUnauthorized, payload)
	if err != nil {
		return err
	}

	return nil
}

// passwordMatches checks whether a plain-text password matches a hashed password.
func (app *application) passwordMatches(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
