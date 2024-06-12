package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// CodeRequest represents the code to be sent to the AI service.
type CodeRequest struct {
	Code string `json:"code"`
}

// CodeResponse represents the response from the AI service.
type CodeResponse struct {
	Feedback string `json:"feedback"`
}

// aiServiceMock simulates an AI service.
func aiServiceMock(code string) (*CodeResponse, error) {
	// Simulate some AI processing.
	feedback := fmt.Sprintf("AI feedback: Your code is %s!", code)
	return &CodeResponse{Feedback: feedback}, nil
}

func main() {
	// Initialize zerolog
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	mux := http.NewServeMux()

	// Handle POST requests to /code
	mux.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Use a decoder to parse the JSON stream directly.
		decoder := json.NewDecoder(r.Body)
		var req CodeRequest
		if err := decoder.Decode(&req); err != nil {
			log.Error().Err(err).Msg("Failed to decode request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Send the code to the AI service.
		resp, err := aiServiceMock(req.Code)
		if err != nil {
			log.Error().Err(err).Msg("Error calling AI service")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Marshal the response and send it back.
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
	})

	// Create a server with a custom shutdown handler
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go func() {
		log.Info().Msg("Server listening on port 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Error starting server")
		}
	}()

	// Wait for a signal to shut down
	<-stopChan

	// Create a 5-second timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	log.Info().Msg("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}
	log.Info().Msg("Server shut down successfully.")
}
