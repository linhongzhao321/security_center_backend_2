package main

import (
	"github.com/apex/log"

	"interfaces"
)

func main() {
	cmd := interfaces.Root()
	err := cmd.Execute()
	if err != nil {
		log.WithError(err).Error("cmd.Execute error")
	}
}
