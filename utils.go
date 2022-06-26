package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/afero"
)

var obtainPathInfo = func(filepath string, flags Flags) (fs.FileInfo, error) {
	var finalPath string
	// if it is an absolute path
	if strings.HasPrefix(filepath, "/") {
		finalPath = filepath
	} else {
		var homePath, err = getHomePath(flags)
		if err != nil {
			return nil, err
		}
		finalPath = path.Join(homePath, filepath)
	}
	return AppFs.Stat(finalPath)
}

var exitIf = func(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var getHomePath = func(flags Flags) (string, error) {
	var homeDir string
	var err error
	if len(flags.homePath) == 0 {
		homeDir, err = os.UserHomeDir()
	} else {
		homeDir = flags.homePath
	}
	if err != nil {
		return "", errors.New("cannot get home directory for unknown reasons. Do you want to provide a home path by --home-path or -H instead?")
	}
	return homeDir, nil
}

var checkHomePath = func(flags Flags) error {
	if len(flags.homePath) != 0 {
		var _, statErr = AppFs.Stat(flags.homePath)
		if statErr != nil {
			return errors.New("The path " + flags.homePath + " does not make too much sense")
		}
	}
	return nil
}

var resolve = func(filepath string, flags Flags) string {
	if strings.HasPrefix(filepath, "/") {
		return filepath
	} else {
		return path.Join(flags.homePath, filepath)
	}
}

var readTextFromFile = func(filepath string) (string, afero.File, error) {
	var file, err = AppFs.Open(filepath)
	if err != nil {
		return "", nil, err
	}
	var text, readErr = ioutil.ReadAll(file)
	return string(text), file, readErr
}
