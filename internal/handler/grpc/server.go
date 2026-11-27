package grpc

import (
	"context"
	"errors"
	"math"
	"sort"

	moviesv1 "github.com/leonilkuz-13/GoCore/api/proto"
	"github.com/leonilkuz-13/GoCore/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MovieService interface {
	LoadMovies(ctx context.Context, from, to int) (domain.LoadMoviesResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type Server struct {
	moviesv1.UnimplementedMovieServiceServer
	service MovieService
}

func NewServer(service MovieService) *Server {
	return &Server{service: service}
}

func (s *Server) LoadMovies(ctx context.Context, request *moviesv1.LoadMoviesRequest) (*moviesv1.LoadMoviesResponse, error) {
	if request.From < 1 {
		return nil, status.Error(codes.InvalidArgument, "from must be at least 1")
	}
	if request.To < request.From {
		return nil, status.Error(codes.InvalidArgument, "to must be greater than or equal to from")
	}
	if request.From > int64(math.MaxInt) || request.To > int64(math.MaxInt) {
		return nil, status.Error(codes.InvalidArgument, "movie ID exceeds the supported range")
	}

	result, err := s.service.LoadMovies(ctx, int(request.From), int(request.To))
	if err != nil {
		return nil, serviceError(err)
	}

	results := make([]*moviesv1.MovieResult, 0, len(result.Movies)+len(result.Errors))
	for _, movie := range result.Movies {
		results = append(results, &moviesv1.MovieResult{
			Id:    int64(movie.ID),
			Movie: newProtoMovie(movie),
		})
	}
	for _, movieError := range result.Errors {
		results = append(results, &moviesv1.MovieResult{
			Id:    int64(movieError.ID),
			Error: movieError.Error,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})

	return &moviesv1.LoadMoviesResponse{Results: results}, nil
}

func (s *Server) GetLastMovie(ctx context.Context, _ *moviesv1.GetLastMovieRequest) (*moviesv1.Movie, error) {
	movie, err := s.service.GetLastMovie(ctx)
	if err != nil {
		return nil, serviceError(err)
	}
	return newProtoMovie(movie), nil
}

func newProtoMovie(movie domain.Movie) *moviesv1.Movie {
	return &moviesv1.Movie{
		Id:       int64(movie.ID),
		Title:    movie.Title,
		Year:     int64(movie.Year),
		Director: movie.Director,
	}
}

func serviceError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "external movie service timeout")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, "invalid movie ID range")
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "movie not found")
	case errors.Is(err, domain.ErrUpstream):
		return status.Error(codes.Unavailable, "external movie service unavailable")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
