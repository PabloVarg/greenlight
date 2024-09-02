package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"greenlight.pvargasb.com/internal/validator"
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
		app.logger.Error(err, nil)
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, target any) error {
	const MB = 1024 * 1024

	r.Body = http.MaxBytesReader(w, r.Body, MB)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
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

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", MB)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	result := qs.Get(key)
	if result == "" {
		return defaultValue
	}

	return result
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	result := qs.Get(key)
	if result == "" {
		return defaultValue
	}

	return strings.Split(result, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	result := qs.Get(key)
	if result == "" {
		return defaultValue
	}

	intResult, err := strconv.Atoi(result)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return intResult
}

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				app.logger.Error(fmt.Errorf("%s", r), nil)
			}
		}()

		fn()
	}()
}
