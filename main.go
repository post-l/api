package main

import "github.com/go-martini/martini"

func main() {
	m := martini.Classic()
	m.Post("/hn/average", hnFileAverage)
	m.Run()
}
