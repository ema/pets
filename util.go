// Copyright (C) 2022 Emanuele Rocca
//
// A bunch of misc helper functions

package main

import (
	"bytes"
	"os/exec"
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
