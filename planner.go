// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// PetsCause conveys the reason behind a given action.
type PetsCause int

const (
	NONE   = iota // no reason at all
	PKG           // required package is missing
	CREATE        // configuration file is missing and needs to be created
	UPDATE        // configuration file differs and needs to be updated
	OWNER         // needs chown()
	MODE          // needs chmod()
	POST          // post-update command
)

func (pc PetsCause) String() string {
	return map[PetsCause]string{
		PKG:    "PACKAGE_INSTALL",
		CREATE: "FILE_CREATE",
		UPDATE: "FILE_UPDATE",
		OWNER:  "OWNER",
		MODE:   "CHMOD",
		POST:   "POST_UPDATE",
	}[pc]
}

// A PetsAction represents something to be done, namely running a certain
// Command. PetsActions exist because of some Trigger, which is a PetsFile.
type PetsAction struct {
	Cause   PetsCause
	Command *exec.Cmd
	Trigger *PetsFile
}

// String representation of a PetsAction
func (pa *PetsAction) String() string {
	if pa.Trigger != nil {
		return fmt.Sprintf("[%s] %s triggered command: '%s'", pa.Cause, pa.Trigger.Source, pa.Command)
	} else {
		return fmt.Sprintf("[%s] triggered command: '%s'", pa.Cause, pa.Command)
	}
}

// Perform executes the Command
func (pa *PetsAction) Perform() error {
	stdout, stderr, err := RunCmd(pa.Command)

	if err != nil {
		log.Printf("[ERROR] running Perform() -> %v\n", err)
	}

	if len(stdout) > 0 {
		log.Printf("[INFO] stdout from Perform() -> %v\n", stdout)
	}

	if len(stderr) > 0 {
		log.Printf("[ERROR] stderr from Perform() -> %v\n", stderr)
	}

	return err
}

// PkgsToInstall returns two values, a boolean and a command. The former is
// true if there are any new packages to install, the latter is the
// distro-specific command to run to install the packages.
func PkgsToInstall(triggers []*PetsFile) (bool, *exec.Cmd) {
	installPkgs := false
	installCmd := InstallCommand()

	for _, trigger := range triggers {
		for _, pkg := range trigger.Pkgs {
			if SliceContains(installCmd.Args, string(pkg)) {
				log.Printf("[DEBUG] %s already marked to be installed\n", pkg)
			} else if pkg.IsInstalled() {
				log.Printf("[DEBUG] %s already installed\n", pkg)
			} else {
				log.Printf("[INFO] %s not installed\n", pkg)
				installCmd.Args = append(installCmd.Args, string(pkg))
				installPkgs = true
			}
		}
	}

	return installPkgs, installCmd
}

// FileToCopy figures out if the given trigger represents a file that needs to
// be updated, and returns the corresponding PetsAction.
func FileToCopy(trigger *PetsFile) *PetsAction {
	cause := trigger.NeedsCopy()

	if cause == NONE {
		return nil
	} else {
		return &PetsAction{
			Cause:   cause,
			Command: NewCmd([]string{"/bin/cp", trigger.Source, trigger.Dest}),
			Trigger: trigger,
		}
	}
}

// Chown returns a chown PetsAction or nil if none is needed.
func Chown(trigger *PetsFile) *PetsAction {
	// Build arg (eg: 'root:staff', 'root', ':staff')
	arg := ""
	var wantUid, wantGid int
	var err error

	if trigger.User != nil {
		arg = trigger.User.Username

		// get the requested uid as integer
		wantUid, err = strconv.Atoi(trigger.User.Uid)
		if err != nil {
			// This should really never ever happen, unless we're
			// running on Windows. :)
			panic(err)
		}
	}

	if trigger.Group != nil {
		arg = fmt.Sprintf("%s:%s", arg, trigger.Group.Name)

		// get the requested gid as integer
		wantGid, err = strconv.Atoi(trigger.Group.Gid)
		if err != nil {
			panic(err)
		}
	}

	if arg == "" {
		// Return immediately if the file had no 'owner' / 'group' directives
		return nil
	}

	// The action to (possibly) perform is a chown of the file.
	action := &PetsAction{
		Cause:   OWNER,
		Command: NewCmd([]string{"/bin/chown", arg, trigger.Dest}),
		Trigger: trigger,
	}

	// stat(2) the destination file to see if a chown is needed
	fileInfo, err := os.Stat(trigger.Dest)
	if os.IsNotExist(err) {
		// If the destination file is not there yet, prepare a chown
		// for later on.
		return action
	}

	stat, _ := fileInfo.Sys().(*syscall.Stat_t)

	if trigger.User != nil && int(stat.Uid) != wantUid {
		log.Printf("[INFO] %s is owned by uid %d instead of %s\n", trigger.Dest, stat.Uid, trigger.User.Username)
		return action
	}

	if trigger.Group != nil && int(stat.Gid) != wantGid {
		log.Printf("[INFO] %s is owned by gid %d instead of %s\n", trigger.Dest, stat.Gid, trigger.Group.Name)
		return action
	}

	log.Printf("[DEBUG] %s is owned by %d:%d already\n", trigger.Dest, stat.Uid, stat.Gid)
	return nil
}

// Chmod returns a chmod PetsAction or nil if none is needed.
func Chmod(trigger *PetsFile) *PetsAction {
	if trigger.Mode == "" {
		// Return immediately if the 'mode' directive was not specified.
		return nil
	}

	// The action to (possibly) perform is a chmod of the file.
	action := &PetsAction{
		Cause:   MODE,
		Command: NewCmd([]string{"/bin/chmod", trigger.Mode, trigger.Dest}),
		Trigger: trigger,
	}

	// stat(2) the destination file to see if a chmod is needed
	fileInfo, err := os.Stat(trigger.Dest)
	if os.IsNotExist(err) {
		// If the destination file is not there yet, prepare a mod
		// for later on.
		return action
	}

	// See if the desired mode and reality differ.
	newMode, err := StringToFileMode(trigger.Mode)
	if err != nil {
		log.Println("[ERROR] unexpected error in Chmod()", err)
		return nil
	}

	oldMode := fileInfo.Mode()

	if oldMode != newMode {
		log.Printf("[INFO] %s is %s instead of %s\n", trigger.Dest, oldMode, newMode)
		return action
	}

	log.Printf("[DEBUG] %s is %s already\n", trigger.Dest, newMode)
	return nil
}

// NewPetsActions is the []PetsFile -> []PetsAction constructor.  Given a slice
// of PetsFile(s), generate a list of PetsActions to perform.
func NewPetsActions(triggers []*PetsFile) []*PetsAction {
	actions := []*PetsAction{}

	// First, install all needed packages. Build a list of all missing package
	// names first, and then install all of them in one go. This is to avoid
	// embarassing things like running in a loop apt install pkg1 ; apt install
	// pkg2 ; apt install pkg3 like some configuration management systems do.
	if installPkgs, installCmd := PkgsToInstall(triggers); installPkgs {
		actions = append(actions, &PetsAction{
			Cause:   PKG,
			Command: installCmd,
		})
	}

	for _, trigger := range triggers {
		actionFired := false

		// Then, figure out which files need to be modified/created.
		if fileAction := FileToCopy(trigger); fileAction != nil {
			actions = append(actions, fileAction)
			actionFired = true
		}

		// Any owner changes needed
		if chown := Chown(trigger); chown != nil {
			actions = append(actions, chown)
			actionFired = true
		}

		// Any mode changes needed
		if chmod := Chmod(trigger); chmod != nil {
			actions = append(actions, chmod)
			actionFired = true
		}

		// Finally, post-update commands
		if trigger.Post != nil && actionFired {
			actions = append(actions, &PetsAction{
				Cause:   POST,
				Command: trigger.Post,
			})
		}
	}

	return actions
}
