package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type InitCmd struct {
	Hostname  string `help:"Generate sample content for a site at this hostname" required:""`
	Directory string `help:"Generate files in this directory" required:""`
}

var hostRe = regexp.MustCompile(`(?:https?://)?([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)`)

func (a *InitCmd) Run(globals *Globals) error {
	matches := hostRe.FindStringSubmatch(a.Hostname)
	var baseURL, brand string
	if matches == nil {
		baseURL = "https://e.g.example.net"
		brand = "Your Brand"
	} else {
		baseURL = fmt.Sprintf("https://%s.%s", matches[2], matches[3])
		brand = strings.ToLower(matches[2])
	}
	err := os.MkdirAll(a.Directory, 0755)
	if err != nil {
		return err
	}
	err = fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		tpl, err := template.New(d.Name()).ParseFS(templateFS, path)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(a.Directory, strings.TrimSuffix(d.Name(), ".tpl")))
		if err != nil {
			return err
		}
		err = tpl.Execute(f, struct {
			BaseURL   string
			BrandName string
		}{
			BaseURL:   baseURL,
			BrandName: brand,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fatal("Failed to generate content: %s", err)
	}
	printSuccess(globals, "Initialized OK")
	return nil
}
