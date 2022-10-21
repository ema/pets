// Copyright (C) 2022 Emanuele Rocca
//
// Pets configuration parser. Walk through a Pets directory and parse
// modelines.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MAXLINES int = 10

func readModelines(path string) ([]string, error) {
	modelines := []string{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scannedLines := 0
	for scanner.Scan() {
		if scannedLines == MAXLINES {
			return modelines, nil
		}

		line := scanner.Text()

		if strings.Contains(line, "pets:") {
			modelines = append(modelines, line)
		}

		scannedLines += 1
	}
	return modelines, nil
}

func parseModeline(line string, pf *PetsFile) error {
	// We just ignore and throw away anything before the 'pets:' modeline
	// identifier
	re, err := regexp.Compile("pets:(.*)")
	if err != nil {
		return err
	}

	// Hope for the best but prepare for the worst. Here's the error object
	// we're gonna return if things go wrong.
	lineError := fmt.Errorf("ERR: invalid pets modeline: %v", line)

	matches := re.FindStringSubmatch(line)

	if len(matches) < 2 {
		return lineError
	}

	fmt.Println("---")

	components := strings.Split(matches[1], " ")
	for i := 0; i < len(components); i++ {
		elem := components[i]

		if len(elem) == 0 {
			continue
		}

		directive, _, _ := strings.Cut(elem, "=")

		if directive == "destfile" {
			fmt.Printf("\tUHUH found destfile -> %v\n", components)
		} else if directive == "package" {
			fmt.Printf("\tUHUH found package -> %v\n", components)
		}
	}

	return nil
}

// This function is called once for each file in the Pets configuration
// directory
func petsFileHandler(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Skip directories
		return nil
	}

	modelines, err := readModelines(path)
	if err != nil {
		return err
	}

	if len(modelines) == 0 {
		// Not a Pets file. We don't take it personal though
		return nil
	}

	fmt.Printf("-> %d pets modelines found in %s\n", len(modelines), path)

	// Instantiate a PetsFile representation. The only thing we know so far
	// is the source path. Every long journey begins with a single step!
	pf := &PetsFile{
		Source: path,
	}

	for _, line := range modelines {
		err := parseModeline(line, pf)
		if err != nil {
			// TODO: Possibly a syntax error, skip the whole file
			// (this very "path")
			return err
		}
	}

	fmt.Println(pf)
	return err
}

func walkDir(directory string) error {
	err := filepath.Walk(directory, petsFileHandler)
	return err
}
