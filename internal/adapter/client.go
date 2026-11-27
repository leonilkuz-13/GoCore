package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/leonilkuz-13/GoCore/internal/domain"
)

type movieDTO struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func (d movieDTO) toDomain() domain.Movie {
	return domain.Movie{
		ID:       d.ID,
		Title:    d.Title,
		Year:     d.Year,
		Director: d.Director,
	}
}

type MovieClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func NewMovieClient(baseURL string, timeout time.Duration) *MovieClient {
	return &MovieClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

func (cl *MovieClient) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	reqCtx, cancel := context.WithTimeout(ctx, cl.timeout)
	defer cancel()

	url := fmt.Sprintf("%s/%d/info.0.json", cl.baseURL, id)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("create request for movie %d: %w", id, err)
	}

	resp, err := cl.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return domain.Movie{}, fmt.Errorf("request movie %d: %w", id, err)
		}
		return domain.Movie{}, fmt.Errorf("request movie %d: %w: %w", id, domain.ErrUpstream, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// OK
	case http.StatusNotFound:
		return domain.Movie{}, fmt.Errorf("movie %d: %w", id, domain.ErrNotFound)
	default:
		return domain.Movie{}, fmt.Errorf("movie %d: unexpected status %d: %w", id, resp.StatusCode, domain.ErrUpstream)
	}

	var dto movieDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return domain.Movie{}, fmt.Errorf("read movie %d response: %w", id, err)
		}
		return domain.Movie{}, fmt.Errorf("decode movie %d: %w: %w", id, domain.ErrUpstream, err)
	}

	return dto.toDomain(), nil
}
