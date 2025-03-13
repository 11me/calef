package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/11me/calef/models"
	"github.com/11me/calef/services"
)

var httpLogger = slog.With("service", "http")

func HandleSubmitPortfolio(svc *services.ControlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var portfolio models.Portfolio

		err := json.NewDecoder(r.Body).Decode(&portfolio)
		if err != nil {
			httpLogger.Error("failed to decode portfolio", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = svc.SubmitPortfolio(r.Context(), &portfolio)
		if err != nil {
			httpLogger.Error("failed to submit portfolio", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleStopPortfolio(svc *services.ControlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid := r.PathValue("id")

		err := svc.StopPortfolio(r.Context(), pid)
		if err != nil {
			httpLogger.Error("failed to stop portfolio", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
