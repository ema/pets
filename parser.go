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

// Because it is important to know when enough is enough.
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

// parseModeline parses a single pets modeline and populates the given PetsFile
// object. The line should something like:
// # pets: destfile=/etc/ssh/sshd_config, owner=root, group=root, mode=0644
func parseModeline(line string, pf *PetsFile) error {
	// We just ignore and throw away anything before the 'pets:' modeline
	// identifier
	re, err := regexp.Compile("pets:(.*)")
	if err != nil {
		return err
	}

	matches := re.FindStringSubmatch(line)

	if len(matches) < 2 {
		// We thought this was a pets modeline -- but then it turned out to be
		// something different, very different indeed.
		return fmt.Errorf("ERROR: invalid pets modeline: %v", line)
	}

	components := strings.Split(matches[1], ",")
	for _, comp := range components {
		// Ignore whitespace
		elem := strings.TrimSpace(comp)
		if len(elem) == 0 || elem == "\t" {
			continue
		}

		keyword, argument, found := strings.Cut(elem, "=")

		// Just in case something bad should happen
		badKeyword := fmt.Errorf("ERROR: invalid keyword/argument '%v'", elem)

		if !found {
			return badKeyword // See? :(
		}

		switch keyword {
		case "destfile":
			pf.AddDest(argument)
		case "owner":
			pf.AddUser(argument)
		case "group":
			pf.AddGroup(argument)
		case "mode":
			pf.AddMode(argument)
		case "package":
			// haha gotcha this one has no setter
			pf.Pkg = argument
		case "pre":
			pf.AddPre(argument)
		case "post":
			pf.AddPost(argument)
		default:
			return badKeyword
		}

		// :)
		fmt.Printf("DEBUG: keyword '%v', argument '%v'\n", keyword, argument)
	}

	return nil
}

func walkDir(directory string) ([]*PetsFile, error) {
	var petsFiles []*PetsFile

	fmt.Printf("INFO: watching configuration directory '%s'\n", directory)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		// This function is called once for each file in the Pets configuration
		// directory
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories
			return nil
		}

		modelines, err := readModelines(path)
		if err != nil {
			// Returning the error we stop parsing all other files too. Debatable
			// whether we want to do that here or not. readModelines should not
			// fail technically, so it's probably fine to do it. Alternatively, we
			// could just log to stderr and return nil like we do later on for
			// syntax errors.
			return err
		}

		if len(modelines) == 0 {
			// Not a Pets file. We don't take it personal though
			return nil
		}

		fmt.Printf("INFO: %d pets modelines found in %s\n", len(modelines), path)

		// Instantiate a PetsFile representation. The only thing we know so far
		// is the source path. Every long journey begins with a single step!
		pf := &PetsFile{
			Source: path,
		}

		for _, line := range modelines {
			err := parseModeline(line, pf)
			if err != nil {
				// Possibly a syntax error, skip the whole file but do not return
				// an error! Otherwise all other files will be skipped too.
				fmt.Println(err) // XXX: log to stderr
				return nil
			}
		}

		if pf.Dest == "" {
			// Destile is a mandatory argument. If we did not find any, consider it an
			// error.
			fmt.Println(fmt.Errorf("ERROR: No 'destfile' directive found in '%s'", path))
			return nil
		}

		fmt.Printf("DEBUG: '%s' syntax OK\n", path)
		petsFiles = append(petsFiles, pf)
		return nil
	})

	return petsFiles, err
}
