package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
)

const timeout = 10 * time.Second

var errToCode = map[errs.ErrorCode]int{
	errs.InvalidValue: http.StatusBadRequest,
	errs.Unauthorized: http.StatusUnauthorized,
	errs.Forbidden:    http.StatusForbidden,
	errs.NotFound:     http.StatusNotFound,
}

type HandlerFunc[T any] func(http.ResponseWriter, *http.Request) (T, error)

func wrapHandler[T any](logger *logger.Logger, handler HandlerFunc[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := handler(w, r)
		if err != nil {
			status := errToStatus(err)

			logger.WithError(err).Error("handle request")
			http.Error(w, http.StatusText(status), status)

			return
		}

		data, err := json.Marshal(res)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

type SeedRequest struct {
	URLs []string `json:"urls"`
}

func (a *App) runHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/seed", wrapHandler(a.logger, a.handleSeed))

	srv := &http.Server{
		Addr:    a.cfg.HTTPAddress,
		Handler: mux,
	}

	errCh := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) handleSeed(w http.ResponseWriter, r *http.Request) (struct{}, error) {
	const maxBody = 100 * 1024

	body := http.MaxBytesReader(w, r.Body, maxBody)

	var req SeedRequest

	err := json.NewDecoder(body).Decode(&req)
	if err != nil {
		return struct{}{}, errs.ErrInvalidValue.WithMessage(err.Error())
	}

	err = a.seed(r.Context(), req.URLs...)
	if err != nil {
		return struct{}{}, errs.WrapError(err)
	}

	return struct{}{}, nil
}

func errToStatus(err error) int {
	tErr, ok := errors.AsType[*errs.Error](err)
	if !ok {
		return http.StatusInternalServerError
	}

	if status, ok := errToCode[tErr.ErrorCode()]; ok {
		return status
	}

	return http.StatusInternalServerError
}
