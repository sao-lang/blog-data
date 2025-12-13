package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func (c *Client) Upsert(index, id string, doc interface{}) error {
	b, _ := json.Marshal(doc)

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), c.ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Upsert error: %s", res.String())
	}

	return nil
}

func (c *Client) Get(index, id string, v interface{}) error {
	res, err := c.ES.Get(index, id)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Get error: %s", res.String())
	}

	return json.NewDecoder(res.Body).Decode(v)
}

func (c *Client) Delete(index, id string) error {
	res, err := c.ES.Delete(index, id)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Delete error: %s", res.String())
	}
	return nil
}

func (c *Client) Bulk(index string, docs []interface{}) error {
	var buf bytes.Buffer

	for _, doc := range docs {
		meta := map[string]interface{}{
			"index": map[string]string{"_index": index},
		}
		m, _ := json.Marshal(meta)
		d, _ := json.Marshal(doc)

		buf.Write(m)
		buf.WriteByte('\n')
		buf.Write(d)
		buf.WriteByte('\n')
	}

	res, err := c.ES.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Bulk error: %s", res.String())
	}

	return nil
}
