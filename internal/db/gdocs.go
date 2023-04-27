package db

import (
	"encoding/json"
	"fmt"
)

const (
	ImageEmbedTag = "image"
	RawEmbedTag   = "raw"
)

type Embed struct {
	N     int    `json:"n"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (em *Embed) UnmarshalJSON(data []byte) error {
	type Nomethods Embed
	var temp = struct {
		*Nomethods
		Value json.RawMessage `json:"value"`
	}{Nomethods: (*Nomethods)(em)}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	switch temp.Type {
	case ImageEmbedTag:
		var img EmbedImage
		if err := json.Unmarshal(temp.Value, &img); err != nil {
			return err
		}
		em.Value = img
	case RawEmbedTag:
		var s string
		if err := json.Unmarshal(temp.Value, &s); err != nil {
			return err
		}
		em.Value = s
	default:
		return fmt.Errorf("unknown embed type tag: %q", temp.Type)
	}
	return nil
}

type EmbedImage struct {
	Path        string `json:"path"`
	Credit      string `json:"credit"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
}
