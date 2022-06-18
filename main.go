package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

func main() {
	const FILE_NAME = "./test"
	var bytes, err = os.ReadFile(FILE_NAME)
	exitIf(err, "Cannot read "+FILE_NAME)
	var text = string(bytes)
	parseFile(text)
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
*/
func parseFile(text string) []Bookmark {
	var lines = strings.Split(text, "\n")
	var bookmarks = make([]Bookmark, len(lines))
	for index, line := range lines {
		fmt.Println("parsing line", index, ":", line)
		var tokens = strings.Split(line, " ")
		// discard anything after the second token
		var firstTwoTokens = tokens[:2]
		var abbreviation, filepath = firstTwoTokens[0], firstTwoTokens[1]
		var info, err = obtainPathInfo(filepath)
		exitIf(err, "Path"+filepath+"does not exist in your file system! If you want to turn off validation, use --no-validate-path or -P")
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
		fmt.Println("bookmark", index, ":", bookmark)
	}
	return bookmarks
}

func obtainPathInfo(filepath string) (fs.FileInfo, error) {
	var finalPath string
	// if it is an absolute path
	if strings.HasPrefix(filepath, "/") {
		finalPath = filepath
	} else {
		var homeDir, err = os.UserHomeDir()
		exitIf(err, "For some reason, the home directory cannot be retrieved. You can explicitly provide a home path by either specifying $HOME or --home-path=<path>")
		finalPath = path.Join(homeDir, filepath)
	}
	return os.Stat(finalPath)
}

func exitIf(err error, message string) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}
