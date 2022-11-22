// Copyright (C) 2022 Emanuele Rocca
//
// A bunch of misc helper functions

package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"testing"
)

// NewCmd is a wrapper for exec.Command. It builds a new *exec.Cmd from a slice
// of strings.
func NewCmd(args []string) *exec.Cmd {
	var cmd *exec.Cmd

	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	return cmd
}

// RunCmd runs the given command and returns two strings, one with stdout and
// one with stderr. The error object returned by cmd.Run() is also returned.
func RunCmd(cmd *exec.Cmd) (string, string, error) {
	var outb bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()

	return outb.String(), errb.String(), err
}

// Sha256 returns the sha256 of the given file. Shocking, I know.
func Sha256(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func StringToFileMode(mode string) (os.FileMode, error) {
	octalMode, err := strconv.ParseInt(mode, 8, 64)
	return os.FileMode(octalMode), err
}

func SliceContains(slice []string, elem string) bool {
	for _, value := range slice {
		if value == elem {
			return true
		}
	}
	return false
}

// Various test helpers
func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("%v != %v", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expecting err to be nil, got %v instead", err)
	}
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expecting an error, got nil instead")
	}
}

func NewTestFile(src, pkg, dest, userName, groupName, mode, pre, post string) (*PetsFile, error) {
	var err error

	p := &PetsFile{
		Source: src,
		Pkgs:   []PetsPackage{PetsPackage(pkg)},
	}

	p.AddDest(dest)

	err = p.AddUser(userName)
	if err != nil {
		return nil, err
	}

	err = p.AddGroup(groupName)
	if err != nil {
		return nil, err
	}

	err = p.AddMode(mode)
	if err != nil {
		return nil, err
	}

	p.AddPre(pre)

	p.AddPost(post)

	return p, nil
}
