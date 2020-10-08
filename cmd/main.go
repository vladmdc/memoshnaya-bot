package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

type RequestBody struct {
	HttpMethod string `json:"httpMethod"`
	Body       string `json:"body"`
}

func main() {
	ctx := context.Background()
	c := createClient(ctx)
	defer c.Close()

	_, _, err := c.Collection("users").Add(ctx, map[string]string{
		"msg": "Vlados",
	})
	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}
}

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "telegram"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}

