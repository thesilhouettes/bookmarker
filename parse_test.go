package main

import (
	"errors"
	"os"
	"path"
	"testing"

	. "github.com/franela/goblin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func addHome(filepath string) string {
	var dir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(dir, filepath)
}

func TestParse(t *testing.T) {
	var g = Goblin(t)

	g.Describe("parseFile works", func() {
		g.Before(func() {
			AppFs = afero.NewMemMapFs()

			AppFs.Mkdir(addHome(".config"), os.ModeDir)
			AppFs.Mkdir(addHome(".config/whatever"), os.ModeDir)
			AppFs.Create(addHome(".config/whatever/conf"))
			log.SetLevel(log.FatalLevel)
		})

		var neededBookmarks = []Bookmark{
			{typ: "dir", path: ".config/", abbreviation: "c"},
			{typ: "file", path: ".config/whatever/conf", abbreviation: "cw"},
		}

		var homePath, _ = os.UserHomeDir()
		var flags = Flags{
			homePath: homePath,
		}

		g.It("parse valid strings that have no comments", func() {
			var res = []struct {
				in   string
				want []Bookmark
			}{
				{
					in:   "c .config/",
					want: []Bookmark{neededBookmarks[0]},
				},
				{
					in:   "cw .config/whatever/conf",
					want: []Bookmark{neededBookmarks[1]},
				},
				{
					in:   "c .config/\ncw .config/whatever/conf",
					want: neededBookmarks,
				},
			}

			for _, pair := range res {
				var out, _ = parseFile(pair.in, flags)
				g.Assert(out).Equal(pair.want)
			}
		})

		g.It("parse valid strings with comments and new lines", func() {
			var res = []struct {
				in   string
				want []Bookmark
			}{
				{
					in:   "#first line is comment\nc .config/",
					want: []Bookmark{neededBookmarks[0]},
				},
				{
					in:   "\ncw .config/whatever/conf\n\n#last line is a comment",
					want: []Bookmark{neededBookmarks[1]},
				},
				{
					in:   "#hi\nc .config/\ncw .config/whatever/conf # between lines\n",
					want: neededBookmarks,
				},
			}

			for _, pair := range res {
				var out, _ = parseFile(pair.in, flags)
				g.Assert(out).Equal(pair.want)
			}
		})

		g.It("can return abbreviation error and filepath error", func() {
			var res = []struct {
				in   string
				want error
			}{
				{
					in:   " /varr",
					want: errors.New("abbreviation is empty"),
				},
				{
					in:   "v /varr",
					want: errors.New("filepath /varr does not exist"),
				},
				{
					in:   "cj .config/jj",
					want: errors.New("filepath .config/jj does not exist"),
				},
			}

			for _, pair := range res {
				var _, err = parseFile(pair.in, flags)
				g.Assert(err).Equal(pair.want)
			}
		})
	})

}
