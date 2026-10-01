package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func fetchMovie(ctx context.Context, id int, timeout time.Duration) (Movie, error) {

	url := fmt.Sprintf("http://homeworksite.site/%d/info.0.json", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Movie{}, fmt.Errorf(">Error: request formatting error: %w", err)
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return Movie{}, fmt.Errorf(">Error: request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Movie{}, fmt.Errorf(">Error: server returned an error, status: %d", resp.StatusCode)
	}

	var movie Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return Movie{}, fmt.Errorf(">Error: malformed JSON: %w ", err)
	}

	return movie, nil
}
