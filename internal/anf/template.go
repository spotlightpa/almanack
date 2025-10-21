package anf

import (
	_ "embed"
	"sync"

	"encoding/json"

	"github.com/spotlightpa/almanack/internal/must"
)

//go:embed sample/article.json
var templateJSON []byte

var templateDoc = sync.OnceValue(func() Article {
	var a Article
	must.Do(json.Unmarshal(templateJSON, &a))
	return a
})
