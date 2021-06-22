package main

import (
	"log"
)

func main() {
	// TODO: use cli arguments and defaults
	a, err := newApp("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	if err = a.setup(); err != nil {
		log.Fatal(err)
	}

	a.run()
}
