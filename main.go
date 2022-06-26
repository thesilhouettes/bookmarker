package main

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var AppFs = afero.NewOsFs()

type Flags struct {
	bookmarkFile string
	// disableValidation bool
	homePath       string
	debug          bool
	editor         string
	shellAliasFile string
}

var parseCommand = func() {
	var flags = Flags{}
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
			checkHomePath(flags)

			// read and parse the input file, returns bookmarks
			// the bookmarks then will be fed to the generators
			var text, file, err = readTextFromFile(flags.bookmarkFile)
			exitIf(err)
			var bms, pathErr = parseFile(string(text), flags)
			exitIf(pathErr)

			defer file.Close()
			// here are the generators
			// a generator is just a function that receives bookmarks array

			var shellAliasErr = generateShellAliases(bms, flags)
			exitIf(shellAliasErr)
			var lfMappingsErr = generateLfMappings(bms, flags)
			exitIf(lfMappingsErr)

			// end of generators

			// so that the user knows the program succeeds
			fmt.Println("Bookmarks has all been generated")
		},
	}
	var homedir, _ = os.UserHomeDir()

	rootCmd.Flags().StringVarP(&flags.homePath, "home-path", "H", homedir, "The home path, uses $HOME if nothing is provided")
	rootCmd.Flags().StringVarP(&flags.bookmarkFile, "bookmark-file", "b", path.Join(homedir, ".config", "bookmarker", "list"), "Input book mark file")
	rootCmd.Flags().StringVarP(&flags.editor, "editor", "e", "", "Editor for shell aliases (it will use $EDITOR if this flag is empty)")
	rootCmd.Flags().StringVarP(&flags.shellAliasFile, "alias-file", "a", path.Join(homedir, ".config", "shell", "aliasrc"), "The filepath for the shell alias file. Remember to source it in your *rc or *profile files")
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

