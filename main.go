package main

import (
  "github.com/simon-engledew/gocmdpev/gopev"
  "gopkg.in/alecthomas/kingpin.v2"
  "github.com/fatih/color"
  "io/ioutil"
  "log"
  "os"
)

var (
  app = kingpin.New("gocmdpev", "A command-line GO Postgres query visualizer (see https://github.com/simon-engledew/gocmdpev).")
)

func main() {
  app.HelpFlag.Short('h')
  app.Version("1.0.1")
  app.VersionFlag.Short('v')
  app.Parse(os.Args[1:])

  buffer, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  // fmt.Println(string(buffer))

  err = gopev.Visualize(color.Output, buffer)

  if err != nil {
    log.Fatalf("%v", err)
  }
}
