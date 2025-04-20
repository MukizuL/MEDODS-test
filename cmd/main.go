package main

import (
	"MEDODS-test/internal/app"
	"log"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
