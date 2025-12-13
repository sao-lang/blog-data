package es

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateIndex(index string, mapping map[string]interface{}) error {
	body, _ := json.Marshal(mapping)

	res, err := c.ES.Indices.Create(index, c.ES.Indices.Create.WithBody(bytes.NewReader(body)))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("CreateIndex error: %s", res.String())
	}
	return nil
}

func (c *Client) IndexExists(index string) (bool, error) {
	res, err := c.ES.Indices.Exists([]string{index})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// 添加别名
func (c *Client) AddAlias(index, alias string) error {
	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{"add": map[string]interface{}{
				"index": index,
				"alias": alias,
			}},
		},
	}
	b, _ := json.Marshal(body)

	res, err := c.ES.Indices.UpdateAliases(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Alias error: %s", res.String())
	}
	return nil
}
