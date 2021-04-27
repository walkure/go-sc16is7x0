package main

import (
	"../../go-sc16is7x0"
	"fmt"
	"github.com/d2r2/go-logger"
	z19 "github.com/eternal-flame-AD/mh-z19"
	"log"
)

// go run cmd/main.go

func main() {
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	conf := &sc16is7x0.Config{Address: 0x48, XtalFreq: 14745600, Baud: 9600}

	dev, err := sc16is7x0.Open(conf)

	if err != nil {
		log.Fatal(err)
	}

	concentration, err := z19.TakeReading(dev)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("co2=%d ppm\n", concentration)
}
