package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/elipatov/web-crawler/pkg/logger"
)

type HandlerFunc[T any] func(http.ResponseWriter, *http.Request) (T, error)

func wrapHandler[T any](logger *logger.Logger, handler HandlerFunc[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := handler(w, r)
		if err != nil {
			logger.Error("handle request", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		data, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Write(data)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

type SeedRequest struct {
	URLs []string `json:"urls"`
}

func (a *App) runHTTPServer() error {
	http.HandleFunc("/seed", wrapHandler(a.logger, a.handleSeed))

	err := http.ListenAndServe(a.cfg.HTTPAddress, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("ListenAndServe: %w", err)
	}

	return nil
}

func (a *App) handleSeed(w http.ResponseWriter, r *http.Request) (struct{}, error) {
	var req SeedRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return struct{}{}, fmt.Errorf("invalid body: %w", err)
	}

	err = a.seed(r.Context(), req.URLs...)
	if err != nil {
		return struct{}{}, fmt.Errorf("seed: %w", err)
	}

	return struct{}{}, nil
}
