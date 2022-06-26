package main

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
must be of format:
[!][abbreviation] [absolute path that can be resolved by your shell]
if the second one ends with a trailing /, then it will be a folder bookmark instead of a file bookmark
c ~/.config/ (will be a folder bookmark)
ac ~/.config/alacritty/alacritty.yml (will be a file bookmark)

if the line starts with a !, that means it is just a normal shell alias

the function also ignores lines with only comment, blank lines and any thing that comes after the format
*/
var parseFile = func(text string, flags Flags) ([]Bookmark, error) {
	var lines = strings.Split(text, "\n")
	var bookmark Bookmark
	// generate 10 lines first
	var bookmarks = make([]Bookmark, 0, 10)
	for index, line := range lines {
		line = strings.Trim(line, " \t\n")
		log.Debugln("parsing line", index, ":", line)
		// leave early if it is just a comment or blank lines
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		} else if strings.HasPrefix(line, "!") {
			var firstSpace = strings.Index(line, " ")
			var asRune = []rune(line)
			if firstSpace == -1 {
				return []Bookmark{}, fmt.Errorf("line %s is missing aliased command", line)
			}
			var firstPart = string(asRune[1:firstSpace])
			var secondPart = string(asRune[firstSpace+1:])

			// valid input
			if len(firstPart) == 0 {
				return []Bookmark{}, fmt.Errorf("line %s is missing abbreviation", line)
			}
			if len(secondPart) == 0 {
				return []Bookmark{}, fmt.Errorf("line %s is missing aliased command", line)
			}

			// if everything passes, generate a bookmark
			bookmark = Bookmark{
				typ:          "shell",
				abbreviation: firstPart,
				path:         secondPart,
			}
		} else {
			var tokens = strings.Split(line, " ")
			if len(tokens) < 2 {
				return nil, fmt.Errorf("not enough arguments for line %s", line)
			}
			// discard anything after the second token
			var firstTwoTokens = tokens[:2]
			var abbreviation, filepath = strings.Trim(firstTwoTokens[0], " \t"), firstTwoTokens[1]

			// check if abbreviation makes sense
			if len(abbreviation) == 0 {
				return nil, errors.New("abbreviation is empty")
			}

			// check filpath
			var info, err = obtainPathInfo(filepath, flags)
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
			bookmark = Bookmark{
				typ:          typ,
				path:         filepath,
				abbreviation: abbreviation,
			}
		}
		bookmarks = append(bookmarks, bookmark)
		log.Debugln("bookmark", index, ":", bookmark)
	}
	return bookmarks, nil
}
