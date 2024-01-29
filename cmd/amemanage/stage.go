package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/wttw/aboutmyemail"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type StageCmd struct {
	Files []string `arg:"" type:"existingfile"`
	Force bool     `help:"Upload files even if tests fail"`
}

func (a *StageCmd) Run(globals *Globals) error {
	if lint(a.Files, globals) {
		if !a.Force {
			fatal("Tests failed, not uploading. Use --force to override")
		}
		printWarning("Tests failed, continuing anyway")
	}
	var buff bytes.Buffer
	mpw := multipart.NewWriter(&buff)
	for _, file := range a.Files {
		ext := filepath.Ext(file)
		if ext != ".md" && ext != ".json" {
			printWarning("Skipping non-markdown, non-json file: %s", file)
			continue
		}
		f, err := os.Open(file)
		if err != nil {
			printError("Failed to open %s, skipping: %s", file, err)
			continue
		}
		if !globals.Quiet {
			blue := color.New(color.FgHiBlue).SprintFunc()
			_, _ = fmt.Fprintf(color.Output, "Uploading %s as %s\n", file, blue(filepath.Base(file)))
		}
		writer, err := mpw.CreateFormFile("filename", filepath.Base(file))
		if err != nil {
			fatal("Failed to create upload: %s", err)
		}
		_, err = io.Copy(writer, f)
		if err != nil {
			fatal("Failed to create upload: %s", err)
		}
	}

	err := mpw.Close()
	if err != nil {
		fatal("Failed to create upload: %s", err)
	}
	client, err := aboutmyemail.New(aboutmyemail.WithServer(globals.Server), aboutmyemail.WithApiKey(globals.ApiKey))
	if err != nil {
		fatal("Failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	response, err := client.ContentPostWithBodyWithResponse(ctx, mpw.FormDataContentType(), bytes.NewReader(buff.Bytes()))
	if err != nil {
		fatal("Upload failed: %s", err)
	}

	if response.StatusCode() == http.StatusOK {
		printSuccess(globals, "Uploaded OK")
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
