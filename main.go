package main

import (
	"log"
	"os"
)

const (
	defaultConfigFile = "config.toml"
)

func main() {
	cfg := defaultConfigFile
	if len(os.Args) > 1 {
		cfg = os.Args[1]
	}

	a, err := newApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err = a.setup(); err != nil {
		log.Fatal(err)
	}

	a.run()
}
