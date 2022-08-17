package main

import (
	"context"
	"fmt"
	"net/url"

	httpbin "github.com/southerncrossedtech/http-bin-org-go-client"
)

func main() {
	host, _ := url.Parse("http://localhost:8085")

	opts := httpbin.Opts{
		Host:  host,
		Debug: true,
		// Version: "v1", // <-- uncomment if versioned paths is needed
		Authorization: httpbin.Authorization{ // <-- uncomment if auth is needed
			// Prefix: "okta", // <-- uncomment if other token prefix is needed
			Token: "some-secure-token",
		},
	}

	// Create new client
	client, err := httpbin.NewClient(&opts)
	if err != nil {
		panic(fmt.Errorf("unable to create httpbin http client: %w", err))
	}

	ctx := context.Background()

	_, err = client.HTTPMethods.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("error calling GET: %w", err))
	}
}
