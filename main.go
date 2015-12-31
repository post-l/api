package main

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/post-l/api/hn"
)

func main() {
	m := martini.Classic()
	m.Post("/hn/average", hnFileAverage)
	m.Run()
}

type hnFileAverageRequest struct {
	fromIsii bool
	toIsii   bool
	interval int
	filename string
	file     multipart.File
}

func hnFileAverage(w http.ResponseWriter, r *http.Request) {
	req, err := parseHnFileAverageRequest(r)
	if err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	defer req.file.Close()
	var sections hn.Sections
	if req.fromIsii {
		sections, err = hn.ParseIsiiFile(req.file)
	} else {
		sections, err = hn.ParseEconomicFile(req.file)
	}
	if err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if err := sections.Average(req.interval); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=average-"+req.filename)
	if req.toIsii {
		sections.WriteIsii(w)
	} else {
		sections.WriteEconomic(w, req.filename)
	}
}

func parseHnFileAverageRequest(r *http.Request) (*hnFileAverageRequest, error) {
	fromIsii, err := strconv.ParseBool(r.FormValue("fromIsii"))
	if err != nil {
		return nil, errors.New("Bad parameter fromIsii")
	}
	toIsii, err := strconv.ParseBool(r.FormValue("toIsii"))
	if err != nil {
		return nil, errors.New("Bad parameter toIsii")
	}
	interval, err := strconv.Atoi(r.FormValue("interval"))
	if interval <= 1 {
		return nil, errors.New("Bad parameter interval")
	}
	file, fHeader, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("Can't get form file: %s", err)
	}
	return &hnFileAverageRequest{
		fromIsii: fromIsii,
		toIsii:   toIsii,
		interval: interval,
		filename: fHeader.Filename,
		file:     file,
	}, nil
}
