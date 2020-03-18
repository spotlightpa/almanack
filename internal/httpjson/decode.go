package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"

	"github.com/spotlightpa/almanack/pkg/errutil"
)

func errorf(status int, format string, v ...interface{}) errutil.Response {
	message := fmt.Sprintf(format, v...)
	return errutil.Response{
		StatusCode: status,
		Message:    message,
		Cause:      errors.New(message),
	}
}

func DecodeRequest(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Thanks to https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			return errorf(http.StatusUnsupportedMediaType, "Content-Type header is not application/json")
		}
	}

	const (
		megabyte = 1 << 20
		maxSize  = 5 * megabyte
	)
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return errorf(http.StatusBadRequest,
				"Request body contains badly-formed JSON (at position %d)",
				syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errorf(http.StatusBadRequest, "Request body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			return errorf(http.StatusBadRequest,
				"Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return errorf(http.StatusBadRequest,
				"Request body contains unknown field %s",
				fieldName)

		case errors.Is(err, io.EOF):
			return errorf(http.StatusBadRequest, "Request body must not be empty")

		case err.Error() == "http: request body too large":
			return errorf(http.StatusRequestEntityTooLarge,
				"Request body must not be larger than %d bytes", maxSize)

		default:
			return errutil.Response{
				StatusCode: http.StatusBadRequest,
				Message:    http.StatusText(http.StatusBadRequest),
				Cause:      err,
			}
		}
	}

	if dec.More() {
		return errorf(http.StatusBadRequest,
			"Request body must only contain a single JSON object")
	}

	return nil
}
