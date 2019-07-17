package main // import "github.com/simon-engledew/gocmdpev"

import (
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/simon-engledew/pev"
	"golang.org/x/sys/unix"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("gocmdpev", "A command-line GO Postgres query visualizer (see https://github.com/simon-engledew/gocmdpev).")
)

func getWidth(fd uintptr) (width uint) {
	ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	if err != nil {
		return 0 // unlimited
	}
	return uint(ws.Col)
}

func main() {
	app.HelpFlag.Short('h')
	app.Version("1.0.0")
	app.VersionFlag.Short('v')
	app.Parse(os.Args[1:])

	err := pev.Visualize(color.Output, os.Stdin, getWidth(os.Stdout.Fd()))

	if err != nil {
		log.Fatalf("%v", err)
	}
}
