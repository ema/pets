// Copyright (C) 2022 Emanuele Rocca
//
// Pets File representation. This is the central data structure of the system:
// it is the in-memory representation of a configuration file (eg: sshd_config)

package main

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

type PetsFile struct {
	Content string
	Pkg     string
	Dest    string
	User    *user.User
	Group   *user.Group
	Mode    os.FileMode
	Pre     *exec.Cmd
	Post    *exec.Cmd
}

func NewPetsFile(content, pkg, dest, userName, groupName, mode, pre, post string) (*PetsFile, error) {
	var err error

	p := &PetsFile{
		Content: content,
		Pkg:     pkg,
		Dest:    dest, // TODO: create dest if missing
	}

	p.User, err = user.Lookup(userName)
	if err != nil {
		// TODO: one day we may add support for creating users
		return nil, err
	}

	p.Group, err = user.LookupGroup(groupName)
	if err != nil {
		// TODO: one day we may add support for creating groups
		return nil, err
	}

	octalMode, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		return nil, err
	}

	p.Mode = os.FileMode(octalMode)

	preArgs := strings.Fields(pre)
	if len(preArgs) > 0 {
		p.Pre = exec.Command(preArgs[0], preArgs[1:]...)
	}

	postArgs := strings.Fields(post)
	if len(postArgs) > 0 {
		p.Post = exec.Command(postArgs[0], postArgs[1:]...)
	}

	return p, nil
}
