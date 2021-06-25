package main

import (
	"log"
	"os"
)

const (
	defaultConfig = "config.toml"
)

func main() {
	cfg := defaultConfig
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
