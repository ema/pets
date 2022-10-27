// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"os/exec"
	"os/user"
	"strings"
)

// PetsFile is the central data structure of the system: it is the in-memory
// representation of a configuration file (eg: sshd_config)
type PetsFile struct {
	Source string
	Pkgs   []PetsPackage
	Dest   string
	User   *user.User
	Group  *user.Group
	// use string instead of os.FileMode to avoid converting back and forth
	Mode string
	Pre  *exec.Cmd
	Post *exec.Cmd
}

func (pf *PetsFile) IsValid(pathErrorOK bool) bool {
	// Check if the specified package(s) exists
	for _, pkg := range pf.Pkgs {
		if !pkg.IsValid() {
			return false
		}
	}

	// Check pre-update validation command
	if !runPre(pf, pathErrorOK) {
		return false
	}

	return true
}

func (pf *PetsFile) AddDest(dest string) {
	// TODO: create dest if missing
	pf.Dest = dest
}

func (pf *PetsFile) AddUser(userName string) error {
	user, err := user.Lookup(userName)
	if err != nil {
		// TODO: one day we may add support for creating users
		return err
	}
	pf.User = user
	return nil
}

func (pf *PetsFile) AddGroup(groupName string) error {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		// TODO: one day we may add support for creating groups
		return err
	}
	pf.Group = group
	return nil
}

func (pf *PetsFile) AddMode(mode string) error {
	_, err := StringToFileMode(mode)
	if err == nil {
		// The specified 'mode' string is valid.
		pf.Mode = mode
	}
	return err
}

func (pf *PetsFile) AddPre(pre string) {
	preArgs := strings.Fields(pre)
	if len(preArgs) > 0 {
		pf.Pre = NewCmd(preArgs)
	}
}

func (pf *PetsFile) AddPost(post string) {
	postArgs := strings.Fields(post)
	if len(postArgs) > 0 {
		pf.Post = NewCmd(postArgs)
	}
}

func NewPetsFile(src, pkg, dest, userName, groupName, mode, pre, post string) (*PetsFile, error) {
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
