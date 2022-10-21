// Copyright (C) 2022 Emanuele Rocca
//
// Pets File representation. This is the central data structure of the system:
// it is the in-memory representation of a configuration file (eg: sshd_config)

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

type PetsFile struct {
	Source string
	Pkg    string
	Dest   string
	User   *user.User
	Group  *user.Group
	Mode   os.FileMode
	Pre    *exec.Cmd
	Post   *exec.Cmd
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
	octalMode, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		return err
	}

	pf.Mode = os.FileMode(octalMode)
	return nil
}

func (pf *PetsFile) AddPre(pre string) {
	preArgs := strings.Fields(pre)
	if len(preArgs) > 0 {
		pf.Pre = exec.Command(preArgs[0], preArgs[1:]...)
	}
}

func (pf *PetsFile) AddPost(post string) {
	postArgs := strings.Fields(post)
	if len(postArgs) > 0 {
		pf.Post = exec.Command(postArgs[0], postArgs[1:]...)
	}
}

// RunPre returns true if the pre-update validation command passes, or if it
// was not specificed at all. The boolean argument pathErrorOK controls whether
// or not we want to fail if the validation command is not around.
func (pf *PetsFile) RunPre(pathErrorOK bool) bool {
	if pf.Pre == nil {
		return true
	}

	// Run 'pre' validation command, append Source filename to
	// arguments.
	// eg: /usr/sbin/sshd -t -f sample_pet/ssh/sshd_config
	pf.Pre.Args = append(pf.Pre.Args, pf.Source)

	err := pf.Pre.Run()

	_, pathError := err.(*fs.PathError)

	if err == nil {
		fmt.Printf("INFO: pre-update command %s successful\n", pf.Pre.Args)
		return true
	} else if pathError && pathErrorOK {
		// The command has failed because the validation command itself is
		// missing. This could be a chicken-and-egg problem: at this stage
		// configuration is not validated yet, hence any "package" directives
		// have not been applied.  Do not consider this as a failure, for now.
		fmt.Printf("INFO: pre-update command %s failed due to PathError. Ignoring for now\n", pf.Pre.Args)
		return true
	} else {
		fmt.Printf("ERROR: pre-update command %s: %s\n", pf.Pre.Args, err)
		return false
	}
}

func (pf *PetsFile) PkgExists() bool {
	// TODO
	return true
}

func NewPetsFile(src, pkg, dest, userName, groupName, mode, pre, post string) (*PetsFile, error) {
	var err error

	p := &PetsFile{
		Source: src,
		Pkg:    pkg,
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
