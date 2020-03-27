package httpjson

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get(ctx context.Context, c *http.Client, url string, v interface{}) error {
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

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bad response: %d %s", resp.StatusCode, resp.Status)
	}

	if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}
