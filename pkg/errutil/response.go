package errutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Response struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
	Cause      error  `json:"-"`
}

func ResponseFrom(err error) Response {
	var errResp Response
	switch {
	case errors.Is(err, context.Canceled):
		err = Response{
			StatusCode: http.StatusServiceUnavailable,
			Message:    "canceled",
			Cause:      err,
		}
	}
	if !errors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Cause = err
	}
	if errResp.Message == "" {
		errResp.Message = http.StatusText(errResp.StatusCode)
	}
	return errResp
}

func (resp Response) Error() string {
	return fmt.Sprintf("[%d] %v: %q", resp.StatusCode, resp.Cause, resp.Message)
}

func (resp Response) Unwrap() error {
	return resp.Cause
}
