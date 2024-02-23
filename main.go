package main

import (
	"log"
	"os"

	"github.com/ciiim/cloudborad/cmd/ring"
	_ "github.com/ciiim/cloudborad/cmd/ring"
)

func main() {
	if err := ring.App.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
