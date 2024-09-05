package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)

	// Movies
	mux.HandleFunc("GET /v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	mux.HandleFunc("GET /v1/movies/{id}", app.requirePermission("movies:read", app.showMovieHandler))
	mux.HandleFunc("POST /v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	mux.HandleFunc("PUT /v1/movies/{id}", app.requirePermission("movies:write", app.updateMovieHandler))
	mux.HandleFunc("PATCH /v1/movies/{id}", app.requirePermission("movies:write", app.partialUpdateMovieHandler))
	mux.HandleFunc("DELETE /v1/movies/{id}", app.requirePermission("movies:write", app.deleteMovieHandler))

	// Users
	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	mux.HandleFunc("PUT /v1/users/activate", app.activateUserHandler)

	// Tokens
	mux.HandleFunc("POST /v1/tokens/authenticate", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(mux))))
}
