// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
	// Is this a symbolic link or an actual file to be copied?
	Link bool
}

func NewPetsFile() *PetsFile {
	return &PetsFile{
		Source: "",
		Dest:   "",
		Mode:   "",
		Link:   false,
	}
}

// NeedsCopy returns PetsCause UPDATE if Source needs to be copied over Dest,
// CREATE if the Destination file does not exist yet, NONE otherwise.
func (pf *PetsFile) NeedsCopy() PetsCause {
	if pf.Link || pf.Source == "" {
		return NONE
	}

	shaSource, err := Sha256(pf.Source)
	if err != nil {
		log.Printf("[ERROR] cannot determine sha256 of Source file %s: %v\n", pf.Source, err)
		return NONE
	}

	shaDest, err := Sha256(pf.Dest)
	if os.IsNotExist(err) {
		return CREATE
	} else if err != nil {
		log.Printf("[ERROR] cannot determine sha256 of Dest file %s: %v\n", pf.Dest, err)
		return NONE
	}

	if shaSource == shaDest {
		log.Printf("[DEBUG] same sha256 for %s and %s: %s\n", pf.Source, pf.Dest, shaSource)
		return NONE
	}

	log.Printf("[DEBUG] sha256[%s]=%s != sha256[%s]=%s\n", pf.Source, shaSource, pf.Dest, shaDest)
	return UPDATE
}

// NeedsLink returns PetsCause LINK if a symbolic link using Source as TARGET
// and Dest as LINK_NAME needs to be created. See ln(1) for the most confusing
// terminology.
func (pf *PetsFile) NeedsLink() PetsCause {
	if !pf.Link || pf.Source == "" || pf.Dest == "" {
		return NONE
	}

	fi, err := os.Lstat(pf.Dest)

	if os.IsNotExist(err) {
		// Dest does not exist yet. Happy path, we are gonna create it!
		return LINK
	}

	if err != nil {
		// There was an error calling lstat, putting all my money on
		// permission denied.
		log.Printf("[ERROR] cannot lstat Dest file %s: %v\n", pf.Dest, err)
		return NONE
	}

	// We are here because Dest already exists and lstat succeeded. At this
	// point there are two options:
	// (1) Dest is already a link to Source \o/
	// (2) Dest is a file, or a directory, or a link to something else /o\
	//
	// In any case there is no action to take, but let's come up with a valid
	// excuse for not doing anything.

	// Easy case first: Dest exists and it is not a symlink
	if fi.Mode()&os.ModeSymlink == 0 {
		log.Printf("[ERROR] %s already exists\n", pf.Dest)
		return NONE
	}

	// Dest is a symlink
	path, err := filepath.EvalSymlinks(pf.Dest)

	if err != nil {
		log.Printf("[ERROR] cannot EvalSymlinks() Dest file %s: %v\n", pf.Dest, err)
	} else if pf.Source == path {
		// Happy path
		log.Printf("[DEBUG] %s is a symlink to %s already\n", pf.Dest, pf.Source)
	} else {
		log.Printf("[ERROR] %s is a symlink to %s instead of %s\n", pf.Dest, path, pf.Source)
	}
	return NONE
}

func (pf *PetsFile) IsValid(pathErrorOK bool) bool {
	// Check if the specified package(s) exists
	for _, pkg := range pf.Pkgs {
		if !pkg.IsValid() {
			return false
		}
	}

	// Check pre-update validation command if the file has changed.
	if pf.NeedsCopy() != NONE && !runPre(pf, pathErrorOK) {
		return false
	}

	return true
}

func (pf *PetsFile) AddDest(dest string) {
	// TODO: create dest if missing
	pf.Dest = dest
}

func (pf *PetsFile) AddLink(dest string) {
	pf.Dest = dest
	pf.Link = true
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
