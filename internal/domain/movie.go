package domain

type Movie struct {
	ID       int
	Title    string
	Year     int
	Director string
}

type MovieError struct {
	ID    int
	Error string
}

type LoadMoviesResult struct {
	Movies []Movie
	Errors []MovieError
}
