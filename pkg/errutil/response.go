package errutil

import (
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
	if !errors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Cause = err
	}
	return errResp
}

func (resp Response) Error() string {
	return fmt.Sprintf("[%d] %v: %q", resp.StatusCode, resp.Cause, resp.Message)
}

func (resp Response) Unwrap() error {
	return resp.Cause
}
