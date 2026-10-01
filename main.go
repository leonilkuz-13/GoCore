package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func LoadFlags() Flags {
	from := flag.Int("from", -1, "ID of the first movie (mandatory)")
	to := flag.Int("to", -1, "id of the last movie (mandatory)")
	workers := flag.Int("workers", 10, "number of workers in the worker pool")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP request timeout")

	flag.Parse()

	if *from == -1 {
		fmt.Fprintln(os.Stderr, ">Error: the --from flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *to == -1 {
		fmt.Fprintln(os.Stderr, ">Error: the --to flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *from < 0 {
		fmt.Fprintln(os.Stderr, ">Error: the --from movie ID cannot be negative")
		os.Exit(1)
	}

	if *to < 0 {
		fmt.Fprintln(os.Stderr, ">Error: the --to movie ID cannot be negative")
		os.Exit(1)
	}

	if *to < *from {
		fmt.Fprintln(os.Stderr, ">Error: the --to flag cannot be less than --from")
		os.Exit(1)
	}

	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, ">Error: the number of workers must be greater than 0")
		os.Exit(1)
	}

	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, ">Error: timeout must be greater than 0")
		os.Exit(1)
	}

	return Flags{
		From:    *from,
		To:      *to,
		Workers: *workers,
		Timeout: *timeout,
	}
}

func worker(ctx context.Context, result chan<- Result, jobs <-chan int, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		movie, err := fetchMovie(ctx, job, timeout)
		res := Result{
			MovieID: job,
			Movie:   movie,
			Err:     err,
		}
		result <- res
	}
}

func main() {
	flags := LoadFlags()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer cancel()
	var wg sync.WaitGroup

	jobs := make(chan int, flags.Workers)
	results := make(chan Result, flags.Workers)

	for i := 1; i <= flags.Workers; i++ {
		wg.Add(1)
		go worker(ctx, results, jobs, flags.Timeout, &wg)
	}

	go func() {
		for i := flags.From; i <= flags.To; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for mov := range results {
		if mov.Err != nil {
			fmt.Printf("> Error during ID %d: %v\n", mov.MovieID, mov.Err)
			continue
		}

		fmt.Printf("%d - %s - %d - %s\n", mov.Movie.ID, mov.Movie.Title, mov.Movie.Year, mov.Movie.Director)
	}
}
