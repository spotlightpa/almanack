package anf

import (
	_ "embed"

	"encoding/json"
)

//go:embed sample/article.json
var templateJSON []byte

var templateDoc = func() Article {
	var a Article
	if err := json.Unmarshal(templateJSON, &a); err != nil {
		panic(err)
	}
	return a
}()
