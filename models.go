package main

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type Result struct {
	MovieID int
	Movie   Movie
	Err     error
}
