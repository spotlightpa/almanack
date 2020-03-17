package errutil

import (
	"errors"
	"fmt"
	"net/http"
)

type Response struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
	Log        string `json:"-"`
}

func ResponseFrom(err error) Response {
	var errResp Response
	if !errors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Log = err.Error()
	}
	return errResp
}

func (resp Response) Error() string {
	return fmt.Sprintf("[%d] %s: %q", resp.StatusCode, resp.Log, resp.Message)
}
