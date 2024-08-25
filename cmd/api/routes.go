package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)

	// Movies
	mux.HandleFunc("GET /v1/movies", app.listMoviesHandler)
	mux.HandleFunc("GET /v1/movies/{id}", app.showMovieHandler)
	mux.HandleFunc("POST /v1/movies", app.createMovieHandler)
	mux.HandleFunc("PUT /v1/movies/{id}", app.updateMovieHandler)
	mux.HandleFunc("PATCH /v1/movies/{id}", app.partialUpdateMovieHandler)
	mux.HandleFunc("DELETE /v1/movies/{id}", app.deleteMovieHandler)

	return app.recoverPanic(app.rateLimit(mux))
}
