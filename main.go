package main

import (
	"github.com/rebuy-de/kubernetes-hpaa/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
