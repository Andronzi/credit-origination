package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SchemaRegistryClient struct {
	baseURL string
}

func NewSchemaRegistryClient(url string) *SchemaRegistryClient {
	return &SchemaRegistryClient{baseURL: url}
}

func (c *SchemaRegistryClient) GetSchemaID(subject string, schema string) (int, error) {
	reqBody := fmt.Sprintf(`{"schema": %q}`, schema)
	res, err := http.Post(
		fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, subject),
		"application/vnd.schemaregistry.v1+json",
		bytes.NewBufferString(reqBody),
	)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var result struct{ ID int }
	json.Unmarshal(body, &result)
	return result.ID, nil
}
