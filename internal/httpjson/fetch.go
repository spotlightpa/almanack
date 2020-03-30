package httpjson

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get(ctx context.Context, c *http.Client, url string, v interface{}, acceptStatuses ...int) error {
	if c == nil {
		c = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = StatusCheck(resp, acceptStatuses...); err != nil {
		return err
	}

	if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

var goodStatuses = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
}

func StatusCheck(resp *http.Response, acceptStatuses ...int) error {
	if len(acceptStatuses) == 0 {
		acceptStatuses = goodStatuses
	}
	for _, code := range acceptStatuses {
		if resp.StatusCode == code {
			return nil
		}
	}

	return fmt.Errorf("unexpected status: %s", resp.Status)
}
