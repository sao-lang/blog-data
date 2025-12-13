package es

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type SearchQuery struct {
	Query map[string]interface{}   `json:"query"`
	Sort  []map[string]interface{} `json:"sort,omitempty"`
	From  int                      `json:"from"`
	Size  int                      `json:"size"`
	Aggs  map[string]interface{}   `json:"aggs,omitempty"`
}

func (c *Client) Search(index string, q SearchQuery) (map[string]interface{}, error) {
	body, _ := json.Marshal(q)

	res, err := c.ES.Search(
		c.ES.Search.WithIndex(index),
		c.ES.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
