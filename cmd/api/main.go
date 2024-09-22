package main

import (
	"fmt"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/server"
)

func main() {

	config.LoadConfig()

	srv := server.NewServer()

	err := srv.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
