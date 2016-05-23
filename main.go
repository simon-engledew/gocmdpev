package main


import (
  "./gopev"
  "io/ioutil"
  "log"
  "os"
)

func main() {
  buffer, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  // fmt.Println(string(buffer))

  err = gopev.Visualize(os.Stdout, buffer)

  if err != nil {
    log.Fatalf("%v", err)
  }
}
