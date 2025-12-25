package main

import (
	"log"

	"github.ibm.com/soub4i/lazystripe/internal/config"
	"github.ibm.com/soub4i/lazystripe/internal/ui"
)

func main() {

	cfg := config.Load()
	if err := ui.Run(cfg.APIKey); err != nil {
		log.Fatal(err)
	}

}
