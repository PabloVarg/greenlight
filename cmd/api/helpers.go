package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) readIDParam(r *http.Request, name string) (int, error) {
	id, err := app.IntPathValue(r, name)
	if err != nil {
		return 0, errors.New("Error reading ID from URL")
	}

	if id < 1 {
		return 0, errors.New("Invalid ID parameter")
	}

	return id, nil
}

func (app *application) IntPathValue(r *http.Request, name string) (int, error) {
	num, err := strconv.Atoi(r.PathValue(name))
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error reading %s from URL", name))
	}

	return num, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	res, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		app.logger.Println(err)
		return errors.New("Error encoding data")
	}

	res = append(res, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
	return nil
}
