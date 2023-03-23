package db

type Embed struct {
	N     int    `json:"n"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// TODO: UnmarshalJSON

type EmbedImage struct {
	Path        string `json:"path"`
	Credit      string `json:"credit"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
}
