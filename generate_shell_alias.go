package main

import (
	"errors"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
)
var generateShellAliases = func(bms []Bookmark, flags Flags) error {
	var editor string
	if len(flags.editor) == 0 {

		var _editor, exists = os.LookupEnv("EDITOR")
		if !exists {
			return errors.New("$EDITOR variable does not exist")
		}
		editor = _editor
	} else {
		editor = flags.editor
	}
	var aliasFile, err = AppFs.Create(flags.shellAliasFile)
	if err != nil {
		return err
	}
	defer aliasFile.Close()
	var lines = ""
	for _, bm := range bms {
		var line string
		if bm.typ == "dir" {
			line = fmt.Sprintf("alias c%s='cd %s'\n", bm.abbreviation, resolve(bm.path, flags))
		} else if bm.typ == "file" {
			line = fmt.Sprintf("alias cf%s='%s %s'\n", bm.abbreviation, editor, resolve(bm.path, flags))
		} else {
			line = fmt.Sprintf("alias %s='%s'\n", bm.abbreviation, bm.path)
		}
		log.Debugln(line)
		lines += line
	}
	var _, writeErr = aliasFile.WriteString(lines)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
