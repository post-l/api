package main

import (
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/post-l/api/hn"
)

func main() {
	m := martini.Classic()
	m.Post("/hn/isii/average", hnIsiiFileAverage)
	m.Run()
}

func hnIsiiFileAverage(w http.ResponseWriter, r *http.Request) {
	interval, err := strconv.Atoi(r.FormValue("interval"))
	if interval <= 1 {
		renderJSON(w, http.StatusBadRequest, Error{"Bad parameter interval"})
		return
	}
	f, fHeader, err := r.FormFile("file")
	if err != nil {
		renderJSON(w, http.StatusBadRequest, Error{"Can't get form file: " + err.Error()})
		return
	}
	defer f.Close()
	sections, err := hn.ParseIsiiFile(f)
	if err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if err := sections.Average(interval); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=average.txt")
	sections.WriteStd(w, fHeader.Filename)
}
