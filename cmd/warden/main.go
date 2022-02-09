package main

import (
	log "github.com/eqto/go-logger"
	"github.com/eqto/service"
	"github.com/eqto/warden/internal/warden"
)

func main() {
	defer service.HandlePanic()
	service.OnPanic(log.Fatal)

	if e := service.Run(warden.Run); e != nil {
		log.E(e)
	}

}
