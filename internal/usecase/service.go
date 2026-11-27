package usecase

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"

	"github.com/leonilkuz-13/GoCore/internal/domain"
)

type MovieFetcher interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUsecase struct {
	fetcher        MovieFetcher
	maxConcurrency int
}

func NewMovieUsecase(fetcher MovieFetcher, maxConcurrency int) *MovieUsecase {
	return &MovieUsecase{
		fetcher:        fetcher,
		maxConcurrency: maxConcurrency,
	}
}

type movieResult struct {
	id    int
	movie domain.Movie
	err   error
}

func (u *MovieUsecase) worker(ctx context.Context, jobs <-chan int, results chan<- movieResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range jobs {
		if ctx.Err() != nil {
			return // так ошибку может записать то все-таки ? нахуя этот err, если туда ничего не записывать ?
		}
		movie, err := u.fetcher.GetMovie(ctx, id)
		results <- movieResult{
			id:    id,
			movie: movie,
			err:   err,
		}
	}
}

func (u *MovieUsecase) LoadMovies(ctx context.Context, from, to int) (domain.LoadMoviesResult, error) {
	if from < 1 || to < from {
		return domain.LoadMoviesResult{}, fmt.Errorf("invalid range [%d; %d]: %w", from, to, domain.ErrInvalidArgument)
	}

	workers := u.maxConcurrency
	jobs := make(chan int, workers)
	results := make(chan movieResult, workers)

	var wg sync.WaitGroup
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go u.worker(ctx, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for i := from; i <= to; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
				// OK
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	out := domain.LoadMoviesResult{
		Movies: make([]domain.Movie, 0, to-from+1),
		Errors: make([]domain.MovieError, 0),
	}

	for res := range results {
		if res.err != nil {
			out.Errors = append(out.Errors, domain.MovieError{
				ID:    res.id,
				Error: res.err.Error(),
			})
			continue
		}
		out.Movies = append(out.Movies, res.movie)
	}

	if err := ctx.Err(); err != nil {
		return domain.LoadMoviesResult{}, err
	}

	slices.SortFunc(out.Movies, func(a, b domain.Movie) int {
		return cmp.Compare(a.ID, b.ID)
	})
	slices.SortFunc(out.Errors, func(a, b domain.MovieError) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return out, nil
}

func (u *MovieUsecase) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	if err := ctx.Err(); err != nil {
		return domain.Movie{}, err
	}

	lastMovie, err := u.fetcher.GetMovie(ctx, 1)
	if err != nil {
		return domain.Movie{}, err
	}

	low := 1
	high := 2

	for {
		movie, err := u.fetcher.GetMovie(ctx, high)
		if err == nil {
			low = high
			lastMovie = movie
			if high > math.MaxInt/2 {
				return domain.Movie{}, fmt.Errorf("movie id upper bound overflow")
			}
			high *= 2
			continue
		}
		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		return domain.Movie{}, err
	}

	for low+1 < high {
		mid := low + (high-low)/2

		movie, err := u.fetcher.GetMovie(ctx, mid)
		if err == nil {
			low = mid
			lastMovie = movie
			continue
		}
		if errors.Is(err, domain.ErrNotFound) {
			high = mid
			continue
		}
		return domain.Movie{}, err
	}

	return lastMovie, nil
}
