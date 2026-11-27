package http

import "github.com/leonilkuz-13/GoCore/internal/domain"

type MovieResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type MovieErrorResponse struct {
	ID    int    `json:"id"`
	Error string `json:"error"`
}

type LoadMoviesResponse struct {
	Movies []MovieResponse      `json:"movies"`
	Errors []MovieErrorResponse `json:"errors"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toMovieResponse(m domain.Movie) MovieResponse {
	return MovieResponse{
		ID:       m.ID,
		Title:    m.Title,
		Year:     m.Year,
		Director: m.Director,
	}
}

func toLoadMoviesResponse(res domain.LoadMoviesResult) LoadMoviesResponse {
	movies := make([]MovieResponse, 0, len(res.Movies))
	for _, m := range res.Movies {
		movies = append(movies, toMovieResponse(m))
	}

	errs := make([]MovieErrorResponse, 0, len(res.Errors))
	for _, e := range res.Errors {
		errs = append(errs, MovieErrorResponse{
			ID:    e.ID,
			Error: e.Error,
		})
	}

	return LoadMoviesResponse{
		Movies: movies,
		Errors: errs,
	}
}
