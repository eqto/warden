package main

import (
	log "github.com/eqto/go-logger"
	"github.com/eqto/service"
	"github.com/eqto/warden/internal/warden"
)

func main() {
	defer service.HandlePanic()
	service.OnPanic(log.E)
	log.SetFile(`warden.log`)

	if e := service.Run(warden.Run); e != nil {
		panic(e)
	}

}
