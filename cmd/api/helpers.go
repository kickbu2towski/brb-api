package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var typeError *json.UnmarshalTypeError
		var invalidMarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed json (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains malformed json")
		case errors.As(err, &typeError):
			if typeError.Field != "" {
				return fmt.Errorf("body contains incorrect type for field %s (at character %d)", typeError.Field, typeError.Offset)
			}
			return fmt.Errorf("body contains incorrect type for a field (at character %d)", typeError.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body is empty")
		case errors.As(err, &invalidMarshalError):
			panic("readJSON: incorrect destination when decoding")
		default:
			return err
		}
	}

	/* TODO:
	   - throw error if the body size exceeds a certain limit (say 1MB)
	   - throw error if the body contains extra fields that aren't in the input
		- throw error if the body contains extra values. like this -> '{"name": "kick"} :lol'
	*/

	return nil
}
