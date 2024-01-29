package main

import (
	"context"
	"github.com/wttw/aboutmyemail"
	"net/http"
	"time"
)

type PublishCmd struct {
}

func (a *PublishCmd) Run(globals *Globals) error {
	client, err := aboutmyemail.New(aboutmyemail.WithServer(globals.Server), aboutmyemail.WithApiKey(globals.ApiKey))
	if err != nil {
		fatal("Failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.StylePublishWithResponse(ctx)
	if err != nil {
		fatal("Failed to publish: %s", err)
	}
	if response.StatusCode() == http.StatusOK {
		printSuccess(globals, "Published OK")
	} else {
		printWarning("Server responded with %s", response.HTTPResponse.Status)
		if response.JSON400 != nil {
			printWarning("%s", response.JSON400.Message)
		} else if response.JSON500 != nil {
			printError("%s", response.JSON500.Message)
		} else {
			printError("%s", response.Body)
		}
	}
	return nil
}
