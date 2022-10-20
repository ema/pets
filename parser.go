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
	"strings"
)

func readModelines(path string) error {
	maxLines := 10

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if maxLines == 0 {
			return nil
		}

		line := scanner.Text()

		if strings.Contains(line, "pets:") {
			fmt.Println(line)
		}

		maxLines -= 1
	}
	return nil
}

func handler(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Skip directories
		return nil
	}

	fmt.Println("***", path, info.Size(), "***")
	readModelines(path)

	// TODO: figure out if the file at path contains any pets modelines

	return nil
}

func walkDir(directory string) error {
	err := filepath.Walk(directory, handler)
	return err
}
