package app

import (
	"MEDODS-test/internal/dto"
	"MEDODS-test/internal/errs"
	"MEDODS-test/internal/helpers"
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (app *Application) CreateToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	sub := r.URL.Query().Get("sub")
	if sub == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, dto.ErrorResponse{Err: http.StatusText(http.StatusBadRequest)})
		return
	}

	tokens, err := app.CreateJWT(ctx, sub, r.RemoteAddr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tokens)
}

func (app *Application) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var req dto.RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, dto.ErrorResponse{Err: http.StatusText(http.StatusBadRequest)})
		return
	}

	tokens, err := app.RefreshTokens(ctx, &req, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, errs.ErrInternalServerError) {
			helpers.WriteJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
			return
		}

		app.logger.Error("Error in RefreshToken", zap.Error(err))
		helpers.WriteJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Err: http.StatusText(http.StatusUnauthorized)})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tokens)
}
