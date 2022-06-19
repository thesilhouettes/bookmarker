package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var AppFs = afero.NewOsFs()

type Flags struct {
	bookmarkFile string
	// disableValidation bool
	homePath string
	debug    bool
	editor   string
}

var flags = Flags{}

var parseCommand = func() {
	var rootCmd = &cobra.Command{
		Use:   "bm",
		Short: "bookmarker -- or file shortcuts",
		Long:  "Input a bookmark file and this program will add them to various applications. \n\nEach line in a the input file looks like this:\n\n[suffix] [path]\n\nFor example, if you have `v .vimrc`, then the program will create an alias `cfv` to your shell profile. When you type cfv, your editor will launch ~/.vimrc. You can change for what program it generates for by modifying the source code or providing flags.",
		Run: func(cmd *cobra.Command, args []string) {
			// set log level - can be configured with command line arguments
			log.SetFormatter(&log.TextFormatter{
				ForceColors:            true,
				DisableLevelTruncation: true,
				PadLevelText:           true,
				DisableTimestamp:       true,
			})
			log.SetOutput(os.Stdout)

			// enable debug output
			if flags.debug {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.FatalLevel)
			}

			// check if home path makes sense
			checkHomePath()

			// read and parse the input file, returns bookmarks
			// the bookmarks then will be fed to the generators
			var file, err = AppFs.Open(flags.bookmarkFile)
			defer file.Close()
			exitIf(err)
			var text, readErr = ioutil.ReadAll(file)
			exitIf(readErr)
			var bms, pathErr = parseFile(string(text))
			exitIf(pathErr)

			// here are the generators
			// a generator is just a function that receives bookmarks array

			generateShellAliases(bms)
			generateLfMappings(bms)

			// end of generators

			// so that the user knows the program succeeds
			fmt.Println("Bookmarks has all been generated")
		},
	}
	var homedir, _ = os.UserHomeDir()

	rootCmd.Flags().StringVarP(&flags.homePath, "home-path", "H", homedir, "The home path")
	rootCmd.Flags().StringVarP(&flags.bookmarkFile, "bookmark-file", "b", path.Join(homedir, ".config", "bookmarker", "list"), "Input book mark file")
	rootCmd.Flags().StringVarP(&flags.editor, "editor", "e", "", "Editor for shell aliases (it will use $EDITOR if this flag is empty)")
	rootCmd.Flags().BoolVarP(&flags.debug, "debug", "v", false, "Enable debug output (warning: lots of unnecessary information)")
	// rootCmd.Flags().BoolVarP(&flags.disableValidation, "no-validate-path", "P", false, "Do not check if the paths in the input file exist")

	// parse the command and run the callback
	var parseErr = rootCmd.Execute()
	exitIf(parseErr)
}

func main() {
	parseCommand()
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
		finalPath = path.Join(getHomePath(), filepath)
	}
	return AppFs.Stat(finalPath)
}

var exitIf = func(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var getHomePath = func() string {
	var homeDir string
	var err error
	if len(flags.homePath) == 0 {
		homeDir, err = os.UserHomeDir()
	} else {
		homeDir = flags.homePath
	}
	if err != nil {
		exitIf(errors.New("cannot get home directory for unknown reasons. Do you want to provide a home path by --home-path or -H instead?"))
	}
	return homeDir
}

var checkHomePath = func() {
	if len(flags.homePath) != 0 {
		var _, statErr = AppFs.Stat(flags.homePath)
		if statErr != nil {
			exitIf(errors.New("The path " + flags.homePath + " does not make too much sense"))
		}
	}
}

var generateShellAliases = func(bms []Bookmark) {
	var editor string
	if len(flags.editor) == 0 {

		var _editor, exists = os.LookupEnv("EDITOR")
		if !exists {
			exitIf(errors.New("$EDITOR variable does not exist"))
		}
		editor = _editor
	} else {
		editor = flags.editor
	}
	var aliasFile, err = AppFs.Create(path.Join(flags.homePath, ".config", "shell", "aliasrc"))
	defer aliasFile.Close()
	exitIf(err)
	var lines = ""
	for _, bm := range bms {
		var line string
		if bm.typ == "dir" {
			line = fmt.Sprintf("alias c%s='cd %s'\n", bm.abbreviation, resolve(bm.path))
		} else {
			line = fmt.Sprintf("alias cf%s='%s %s'\n", bm.abbreviation, editor, resolve(bm.path))
		}
		log.Debugln(line)
		lines += line
	}
	var _, writeErr = aliasFile.WriteString(lines)
	exitIf(writeErr)
}

var generateLfMappings = func(bms []Bookmark) {
	const END_GENERATION_STRING = "### End of BOOKMARKER generation"
	const START_GENERATION_STRING = "### Automatically generated by BOOKMARKER ###"
	var lfConfig, err = AppFs.OpenFile(path.Join(flags.homePath, ".config", "lf", "lfrc"), os.O_RDWR, os.ModeAppend)
	defer lfConfig.Close()
	exitIf(err)

	// generate the required strings first
	var lines = START_GENERATION_STRING + "\n\n"
	for _, bm := range bms {
		// we don't need to carry about paths which are not directories
		if bm.typ == "dir" {
			var line = fmt.Sprintf("map g%s cd %s\n", bm.abbreviation, resolve(bm.path))
			log.Debugln(line)
			lines += line
		}
	}
	lines += "\n" + END_GENERATION_STRING + "\n"

	// remove the old section with new section
	var scanner = bufio.NewScanner(lfConfig)
	var success bool = true

	var keptLines = ""
	var insideGenerationSection = false

	// finds the lines which are enclosed by the two comments
	// and not include them
	for {
		success = scanner.Scan()
		if !success {
			exitIf(scanner.Err())
			break
		}
		var line = scanner.Text()

		// omit the lines and check if we are out of it
		if insideGenerationSection {
			if line == END_GENERATION_STRING {
				insideGenerationSection = false
			}
			continue
		}

		// trigger inside generation once the ### line is reached
		if line == START_GENERATION_STRING {
			insideGenerationSection = true
		} else {
			keptLines += line + "\n"
		}
	}

	var allLines = keptLines + lines

	// go back to the beginning and overwrite everything
	lfConfig.Seek(0, 0)

	var _, writeErr = lfConfig.WriteString(allLines)
	exitIf(writeErr)
}

var resolve = func(filepath string) string {
	if strings.HasPrefix(filepath, "/") {
		return filepath
	} else {
		return path.Join(flags.homePath, filepath)
	}
}
