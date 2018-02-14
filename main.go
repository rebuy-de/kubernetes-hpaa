package main

import (
	"github.com/rebuy-de/kubernetes-hpaa/cmd"
	"github.com/rebuy-de/rebuy-go-sdk/cmdutil"
	log "github.com/sirupsen/logrus"
)

func main() {
	defer cmdutil.HandleExit()
	if err := cmd.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
