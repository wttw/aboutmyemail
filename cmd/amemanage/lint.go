package main

import (
	"encoding/json"
	"golang.org/x/text/language/display"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type LintCmd struct {
	Files []string `arg:"" type:"existingfile"`
}

func (a LintCmd) Run(globals *Globals) error {
	if !lint(a.Files, globals) {
		printSuccess(globals, "Linted OK")
	}
	return nil
}

func lint(files []string, _ *Globals) bool {
	failed := false
	expected := map[string]int{}
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			expected[strings.TrimSuffix(d.Name(), ".tpl")] = 1
		}
		return nil
	})
	if err != nil {
		fatal("%s", err)
	}
	splitRe := regexp.MustCompile(`^([a-zA-Z0-9]+)\.([a-zA-Z0-9_-]+)\.md$`)
	languages := map[string]struct{}{}
	for _, file := range files {
		base := filepath.Base(file)
		ext := filepath.Ext(file)
		switch ext {
		case ".json":
			_, ok := expected[base]
			if !ok {
				printWarning("Unexpected file: %s", base)
				failed = true
				continue
			}
			expected[base]++
			if checkJson(file) {
				failed = true
			}
		case ".md":
			baseFile := base
			matches := splitRe.FindStringSubmatch(base)
			if matches != nil {
				baseFile = matches[1] + ".md"
			}
			_, ok := expected[baseFile]
			if !ok {
				printWarning("Unexpected file: %s", base)
				failed = true
				continue
			}
			if matches != nil {
				languages[matches[2]] = struct{}{}
			}
			expected[baseFile]++
		default:
			printWarning("Unexpected file: %s", base)
			failed = true
		}
	}
	supported := map[string]struct{}{}
	for _, tag := range display.Self.Supported.Tags() {
		supported[tag.String()] = struct{}{}
	}
	additionalLanguages := []string{"x-piglatin"}
	for lang := range languages {
		if slices.Contains(additionalLanguages, lang) {
			continue
		}
		_, ok := supported[lang]
		if !ok {
			printWarning("Unexpected language: '%s'", lang)
			failed = true
		}
	}
	return failed
}

func checkJson(file string) bool {
	failed := false
	f, err := os.Open(file)
	if err != nil {
		printError("Failed to open %s: %s", file, err)
		failed = true
	}
	j := map[string]string{}
	dec := json.NewDecoder(f)
	err = dec.Decode(&j)
	if err != nil {
		printError("Failed to parse %s: %s", file, err)
		failed = true
	}
	expectedKeys := []string{"HomeLink", "LogoHTML", "MobileLogoHTML"}
	for key := range j {
		if !slices.Contains(expectedKeys, key) {
			printWarning("Unexpected key in %s: %s", file, key)
			failed = true
		}
	}
	return failed
}
