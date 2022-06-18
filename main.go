package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var AppFs = afero.NewOsFs()

func main() {
	// set log level - can be configured with command line arguments
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	const FILE_NAME = "./test"
	var file, err = AppFs.Open(FILE_NAME)
	exitIf(err, "Cannot read "+FILE_NAME)
	var text, readErr = ioutil.ReadAll(file)
	exitIf(readErr, "The file exists, but the program cannot read it")
	var _, pathErr = parseFile(string(text))
	exitIf(pathErr, "Some paths in the bookmark file do not exist in your file system! If you want to turn off validation, use --no-validate-path or -P")
	fmt.Println("Bookmarks has all been generated")
}

type Bookmark struct {
	typ          string
	path         string
	abbreviation string
}

/*
must be of format:
[abbreviation] [absolute path that can be resolved by your shell]
if the second one ends with a trailing /, then it will be a folder bookmark instead of a file bookmark
c ~/.config/ (will be a folder bookmark)
ac ~/.config/alacritty/alacritty.yml (will be a file bookmark)

the function also ignores lines with only comment, blank lines and any thing that comes after the format
*/
var parseFile = func(text string) ([]Bookmark, error) {
	var lines = strings.Split(text, "\n")
	// generate 10 lines first
	var bookmarks = make([]Bookmark, 0, 10)
	for index, line := range lines {
		log.Debugln("parsing line", index, ":", line)
		// leave early if it is just a comment or blank lines
		if strings.HasPrefix(line, "#") || len(strings.Trim(line, " \t\n")) == 0 {
			continue
		}
		var tokens = strings.Split(line, " ")
		// discard anything after the second token
		var firstTwoTokens = tokens[:2]
		var abbreviation, filepath = strings.Trim(firstTwoTokens[0], " \t"), firstTwoTokens[1]

		// check if abbreviation makes sense
		if len(abbreviation) == 0 {
			return nil, errors.New("abbreviation is empty")
		}

		// check filpath
		var info, err = obtainPathInfo(filepath)
		if err != nil {
			return nil, errors.New("filepath " + filepath + " does not exist")
		}
		// go does not have ternaries. However, since more types may be available, this construct is okay
		var typ string
		if info.IsDir() {
			typ = "dir"
		} else {
			typ = "file"
		}
		// create a bookmark by the information above and append it to bookmarks
		var bookmark = Bookmark{
			typ:          typ,
			path:         filepath,
			abbreviation: abbreviation,
		}
		bookmarks = append(bookmarks, bookmark)
		log.Debugln("bookmark", index, ":", bookmark)
	}
	return bookmarks, nil
}

var obtainPathInfo = func(filepath string) (fs.FileInfo, error) {
	var finalPath string
	// if it is an absolute path
	if strings.HasPrefix(filepath, "/") {
		finalPath = filepath
	} else {
		var homeDir, err = os.UserHomeDir()
		exitIf(err, "For some reason, the home directory cannot be retrieved. You can explicitly provide a home path by either specifying $HOME or --home-path=<path>")
		finalPath = path.Join(homeDir, filepath)
	}
	return AppFs.Stat(finalPath)
}

var exitIf = func(err error, message string) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}
