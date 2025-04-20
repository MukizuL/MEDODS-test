package app

import "github.com/go-chi/chi/v5"

func NewRouter(app *Application) *chi.Mux {
	r := chi.NewRouter()
	r.Use(app.loggerMW)

	r.Get("/api/v1/token", app.CreateToken)
	r.Post("/api/v1/token", app.RefreshToken)

	return r
}
