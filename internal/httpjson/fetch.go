package httpjson

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Get(c *http.Client, url string, v interface{}) error {
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}
