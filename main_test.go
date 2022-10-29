// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestParseFlags(t *testing.T) {
	confDir, debug, dryRun := ParseFlags()
	if len(confDir) == 0 {
		t.Errorf("ParseFlags() returned a empty confDir")
	}
	assertEquals(t, debug, false)
	assertEquals(t, dryRun, false)
}

func TestGetLogFilter(t *testing.T) {
	filter := GetLogFilter(true)
	assertEquals(t, string(filter.MinLevel), "DEBUG")

	filter = GetLogFilter(false)
	assertEquals(t, string(filter.MinLevel), "INFO")
}
