package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"url_reader/config"
	"url_reader/pkg/service"
)

type Api struct {
	cfg config.Config
	svc service.UrlReader
}

type ApiOption func(*Api)

func NewApi(options ...ApiOption) Api {
	api := Api{}
	for _, option := range options {
		option(&api)
	}

	return api
}

func WithConfig(cfg config.Config) ApiOption {
	return func(a *Api) {
		a.cfg = cfg
	}
}

func WithReader(svc service.UrlReader) ApiOption {
	return func(a *Api) {
		a.svc = svc
	}
}

func (a Api) ReadUrls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	data := service.UrlRequests{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		msg := fmt.Sprintf("invalid json data: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, msg)
		return
	}

	// request must contain less then configured requests count
	if len(data.Requests) > a.cfg.MaxUrls {
		writeErrorResponse(w, http.StatusBadRequest, "too many urls in request")
		return
	}

	resp, err := a.svc.Read(r.Context(), data)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot write to stream: %v\n", err)
		return
	}
}
