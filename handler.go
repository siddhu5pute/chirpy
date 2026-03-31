package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"errors"
)


func handlerOne(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Problem while marshaling: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type resError struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, resError{Error: msg})
}

func validateChirp(body string) (string, error){
	if len(body) > 140 {
    return "", errors.New("Chirp is too long")
	}
	cleaned := profanCheck(body)
	return cleaned, nil
}

func profanCheck(msg string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Fields(msg)

	for i, word := range words {
		wordLower := strings.ToLower(word)
		if badWords[wordLower] {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
