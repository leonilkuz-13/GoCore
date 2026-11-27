package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/leonilkuz-13/GoCore/internal/domain"
	httpSwagger "github.com/swaggo/http-swagger"
)

type MovieService interface {
	LoadMovies(ctx context.Context, from, to int) (domain.LoadMoviesResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type Handler struct {
	service MovieService
}

func NewHandler(service MovieService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/load-movies", h.loadMovies)
	mux.HandleFunc("/last-movie", h.lastMovie)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return mux
}

// loadMovies godoc
//
//	@Summary		Load movies by ID range
//	@Description	Loads movies from the external service in the inclusive range [from, to]. Per-movie failures are returned in the corresponding array items.
//	@Tags			movies
//	@Produce		json
//	@Param			from	query		int	true	"First movie ID (at least 1)"	minimum(1)
//	@Param			to		query		int	true	"Last movie ID (greater than or equal to from)"	minimum(1)
//	@Success		200		{array}		MovieResultResponse
//	@Failure		400		{object}	ErrorResponse	"Invalid from or to"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Failure		502		{object}	ErrorResponse	"External movie service error"
//	@Failure		504		{object}	ErrorResponse	"External movie service timeout"
//	@Router			/load-movies [get]
func (h *Handler) loadMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	from, to, err := parseRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.LoadMovies(r.Context(), from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newLoadMoviesResponse(result))
}

// lastMovie godoc
//
//	@Summary		Get the last available movie
//	@Description	Finds and returns the last available movie on the external service.
//	@Tags			movies
//	@Produce		json
//	@Success		200	{object}	MovieResponse
//	@Failure		404	{object}	ErrorResponse	"Movie not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Failure		502	{object}	ErrorResponse	"External movie service error"
//	@Failure		504	{object}	ErrorResponse	"External movie service timeout"
//	@Router			/last-movie [get]
func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	movie, err := h.service.GetLastMovie(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newMovieResponse(movie))
}

func parseRange(r *http.Request) (int, int, error) {
	fromValue := r.URL.Query().Get("from")
	toValue := r.URL.Query().Get("to")
	if fromValue == "" || toValue == "" {
		return 0, 0, errors.New("from and to query parameters are required")
	}

	from, err := strconv.Atoi(fromValue)
	if err != nil {
		return 0, 0, errors.New("from must be an integer")
	}
	to, err := strconv.Atoi(toValue)
	if err != nil {
		return 0, 0, errors.New("to must be an integer")
	}
	if from < 1 {
		return 0, 0, errors.New("from must be at least 1")
	}
	if to < from {
		return 0, 0, errors.New("to must be greater than or equal to from")
	}

	return from, to, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "external movie service timeout")
	case errors.Is(err, domain.ErrInvalidArgument):
		writeError(w, http.StatusBadRequest, "invalid movie ID range")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "movie not found")
	case errors.Is(err, domain.ErrUpstream):
		writeError(w, http.StatusBadGateway, "external movie service error")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", http.MethodGet)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
