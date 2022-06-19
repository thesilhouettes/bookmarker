package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	. "github.com/franela/goblin"
	"github.com/spf13/afero"
)

func TestShellAlias(t *testing.T) {
	var g = Goblin(t)
	AppFs = afero.NewMemMapFs()

	g.Describe("generate correct shell aliases", func() {
		var homeDir, _ = os.UserHomeDir()
		var flags = Flags{
			homePath:       homeDir,
			editor:         "vim",
			shellAliasFile: path.Join(homeDir, ".config", "shell", "aliasrc"),
		}
		g.Before(func() {

			AppFs.MkdirAll(path.Join(flags.homePath, ".config", "shell"), os.ModeDir)
		})

		g.BeforeEach(func() {
			AppFs.Create(path.Join(flags.homePath, ".config", "shell", "aliasrc"))
		})

		g.It("with file type", func() {
			var bookmarks = []Bookmark{
				{
					typ:          "file",
					path:         "weird/file.txt",
					abbreviation: "w",
				},
				{
					typ:          "file",
					path:         "deeply/nested/path/works/fine.conf",
					abbreviation: "wws",
				},
				{
					typ:          "file",
					path:         "/absolute/path/to/nowhere.ini",
					abbreviation: "abs",
				},
			}
			var strs = []string{
				fmt.Sprintf("alias cfw='vim %s/weird/file.txt'", flags.homePath),
				fmt.Sprintf("alias cfwws='vim %s/deeply/nested/path/works/fine.conf'", flags.homePath),
				"alias cfabs='vim /absolute/path/to/nowhere.ini'",
			}
			var want = strings.Join(strs, "\n") + "\n"
			generateShellAliases(bookmarks, flags)
			var text, _, _ = readTextFromFile(path.Join(flags.homePath, ".config", "shell", "aliasrc"))
			g.Assert(text).Equal(want)
		})

		g.It("with dir type", func() {
			var bookmarks = []Bookmark{
				{
					typ:          "dir",
					path:         "weird/folder",
					abbreviation: "w",
				},
				{
					typ:          "dir",
					path:         "deeply/nested/path/works/fine",
					abbreviation: "wws",
				},
				{
					typ:          "dir",
					path:         "/absolute/path/to/nowhere",
					abbreviation: "abs",
				},
			}
			var strs = []string{
				fmt.Sprintf("alias cw='cd %s/weird/folder'", flags.homePath),
				fmt.Sprintf("alias cwws='cd %s/deeply/nested/path/works/fine'", flags.homePath),
				"alias cabs='cd /absolute/path/to/nowhere'",
			}
			var want = strings.Join(strs, "\n") + "\n"
			generateShellAliases(bookmarks, flags)
			var text, _, _ = readTextFromFile(path.Join(flags.homePath, ".config", "shell", "aliasrc"))
			g.Assert(text).Equal(want)
		})

		g.It("with alias shell", func() {
			var bookmarks = []Bookmark{
				{
					typ:          "shell",
					path:         "sudo apt update && sudo apt upgrade",
					abbreviation: "update",
				},
				{
					typ:          "shell",
					path:         "exa -la",
					abbreviation: "ls",
				},
			}
			var strs = []string{
				"alias update='sudo apt update && sudo apt upgrade'",
				"alias ls='exa -la'",
			}
			var want = strings.Join(strs, "\n") + "\n"
			generateShellAliases(bookmarks, flags)
			var text, _, _ = readTextFromFile(path.Join(flags.homePath, ".config", "shell", "aliasrc"))
			g.Assert(text).Equal(want)
		})

	})
}
