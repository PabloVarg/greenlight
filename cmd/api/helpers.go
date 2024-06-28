package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type envelope map[string]any

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

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
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

func (app *application) readJSON(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed JSON at position %d", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains malformed JSON")
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty")
		default:
			return err
		}
	}

	return nil
}
