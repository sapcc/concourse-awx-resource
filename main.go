package main

import (
	"os"

	"github.com/sapcc/concourse-awx-resource/internal/resource"
	"github.com/tbe/resource-framework/log"
	fr "github.com/tbe/resource-framework/resource"
)

func main() {
	r := resource.NewAWXResource()
	handler, err := fr.NewHandler(r)
	if err != nil {
		log.Error("Error creating handler: %s", err)
		os.Exit(1)
	}
	_ = handler.Run()
}
