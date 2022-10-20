// Copyright (C) 2022 Emanuele Rocca
//
// Pets configuration parser. Walk through a Pets directory and parse
// modelines.

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func handler(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Skip directories
		return nil
	}

	fmt.Println(path, info.Size())

	// TODO: figure out if the file at path contains any pets modelines

	return nil
}

func walkDir(directory string) error {
	err := filepath.Walk(directory, handler)
	return err
}
