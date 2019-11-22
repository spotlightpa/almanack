package errutil

import "fmt"

type Response struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
	Log        string `json:"-"`
}

func (resp Response) Error() string {
	return fmt.Sprintf("[%d] %s", resp.StatusCode, resp.Message)
}
