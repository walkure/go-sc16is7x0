# SC16IS7x0 Go Driver

Go interface  for I2C/SPI UART chip [SC16IS740/750/760](https://www.nxp.com/products/peripherals-and-logic/signal-chain/bridges/single-uart-with-ic-bus-spi-interface-64-bs-of-transmit-and-receive-fifos-irda-sir-built-in-support:SC16IS740_750_760).

Tested only SC16IS750 (CJMCU-750).

# dependency

- [d2r2/go-i2c](https://github.com/d2r2/go-i2c)

# implements

- [io.ReadWriteCloser](https://golang.org/pkg/io/#ReadWriteCloser)
- [io.ByteReader](https://golang.org/pkg/io/#ByteReader)
- [io.ByteWriter](https://golang.org/pkg/io/#ByteWriter)

# sample code

use with [eternal-flame-AD/mh-z19](https://github.com/eternal-flame-AD/mh-z19)

``` go
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
```

# refer.
Inspired by [SC16IS750 Python Driver](https://github.com/walkure/SC16IS750) (Original version is [Harri-Renney/SC16IS750](https://github.com/Harri-Renney/SC16IS750)).