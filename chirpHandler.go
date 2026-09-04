package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func customHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type Chirp struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type OKResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		fmt.Println("Decoding error")
		sendError(w, err)
		return
	}

	if len(chirp.Body) > 140 {
		fmt.Println("chirp too long")
		sendError(w, fmt.Errorf("Chirp is too long"))
		return
	}
	sendOK(w, chirp)
}

func sendOK(w http.ResponseWriter, chirp Chirp) {
	res := OKResponse{
		CleanedBody: replaceProfaneWords(chirp.Body),
	}

	byteResponse, err := json.Marshal(res)
	if err != nil {
		fmt.Println("OK could not be sent")
		sendError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(byteResponse)
}

func replaceProfaneWords(input string) string {
	output := input
	profaneWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Fields(input)
	for _, word := range words {
		if profaneWords[strings.ToLower(word)] {
			output = strings.Replace(output, word, "****", 1)
		}
	}
	return output
}
