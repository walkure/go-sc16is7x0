package sc16is7x0

import (
	"errors"
	"github.com/d2r2/go-i2c"
	"time"
)

type Register byte

const (
	regRHR       Register = 0x00 // Receive Holding Register (R)
	regTHR       Register = 0x00 // Transmit Holding Register (W)
	regIER       Register = 0x01 // Interrupt Enable Register (R/W)
	regFCR       Register = 0x02 // FIFO Control Register (W)
	regIIR       Register = 0x02 // Interrupt Identification Register (R)
	regLCR       Register = 0x03 // Line Control Register (R/W)
	regMCR       Register = 0x04 // Modem Control Register (R/W)
	regLSR       Register = 0x05 // Line Status Register (R)
	regMSR       Register = 0x06 // Modem Status Register (R)
	regSPR       Register = 0x07 // Scratchpad Register (R/W)
	regTCR       Register = 0x06 // Transmission Control Register (R/W)
	regTLR       Register = 0x07 // Trigger Level Register (R/W)
	regTXLVL     Register = 0x08 // Transmit FIFO Level Register (R)
	regRXLVL     Register = 0x09 // Receive FIFO Level Register (R)
	regIODIR     Register = 0x0A // I/O pin Direction Register (R/W)
	regIOSTATE   Register = 0x0B // I/O pin States Register (R)
	regIOINTENA  Register = 0x0C // I/O Interrupt Enable Register (R/W)
	regIOCONTROL Register = 0x0E // I/O pins Control Register (R/W)
	regEFCR      Register = 0x0F // Extra Features Register (R/W)

	// -- Special Register Set (Requires LCR[7] = 1 & LCR != 0xBF to use)
	regDLL Register = 0x00 // Divisor Latch LSB (R/W)
	regDLH Register = 0x01 // Divisor Latch MSB (R/W)

	// -- Enhanced Register Set (Requires LCR = 0xBF to use)
	regEFR   Register = 0x02 // Enhanced Feature Register (R/W)
	regXON1  Register = 0x04 // XOn1 (R/W)
	regXON2  Register = 0x05 // XOn2 (R/W)
	regXOFF1 Register = 0x06 // XOff1 (R/W)
	regXOFF2 Register = 0x07 // XOff2 (R/W)
)

type SC16IS7X0 struct {
	dev      *i2c.I2C
	xtalFreq uint32
	timeout  time.Duration
}

const DefaultSize = 8 // Default value for Config.Size

type Config struct {

	//I2C Address
	Address uint8

	//I2C Bus(Default is 1)
	Bus int

	// Desired baudrate
	Baud uint32

	// Frequency of XTAL
	XtalFreq uint32

	// Read and Write Timeout (Default 1sec)
	Timeout time.Duration

	// Size is the number of data bits. If 0, DefaultSize is used.
	Size byte

	// Parity is the bit to use and defaults to ParityNone (no parity bit).
	Parity Parity

	// Number of stop bits to use. Default is 1 (1 stop bit).
	StopBits StopBits
}

type StopBits byte
type Parity byte

const (
	Stop1     StopBits = 1
	Stop1Half StopBits = 15
	Stop2     StopBits = 2
)

const (
	ParityNone  Parity = 'N'
	ParityOdd   Parity = 'O'
	ParityEven  Parity = 'E'
	ParityMark  Parity = 'M' // parity bit is always 1
	ParitySpace Parity = 'S' // parity bit is always 0
)

func (v *SC16IS7X0) readReg(reg Register) (byte, error) {
	return v.dev.ReadRegU8(byte(reg) << 3)
}

func (v *SC16IS7X0) writeReg(reg Register, value byte) error {
	return v.dev.WriteRegU8(byte(reg) << 3, value)
}

func (v *SC16IS7X0) updateRegBit(reg Register, flag byte, value bool) error {

	flags, err := v.readReg(reg)
	if err != nil {
		return err
	}
	if value {
		flags |= (1 << flag)
	} else {
		flags &^= (1 << flag)
	}

	return v.writeReg(reg, flags)
}

func (v *SC16IS7X0) peekRegBit(reg Register, flag byte) (bool, error) {
	current, err := v.readReg(reg)
	if err != nil {
		return false, err
	}
	return current&(1<<flag) != 0, nil
}

func (v *SC16IS7X0) setBaudRate(baud uint32) error {
	clkDiv, err := v.peekRegBit(regMCR, 7)
	if err != nil {
		return err
	}

	prescaler := uint32(1)
	if clkDiv {
		prescaler = 4
	}

	divisor := int((v.xtalFreq / prescaler) / (baud * 16))

	lowerDivisor := byte(divisor & 0xFF)
	higherDivisor := byte((divisor & 0xFF00) >> 8)

	if err = v.updateRegBit(regLCR, 7, true); err != nil {
		return err
	}

	if err = v.writeReg(regDLL, lowerDivisor); err != nil {
		return err
	}

	if err = v.writeReg(regDLH, higherDivisor); err != nil {
		return err
	}

	if err = v.updateRegBit(regLCR, 7, false); err != nil {
		return err
	}

	return nil
}

const testValue = 0xde // test value to write scratchpad register

var ErrorScratchpad error = errors.New("Scratchpad Register value mismatched")

func (v *SC16IS7X0) testChip() error {

	if err := v.writeReg(regSPR, testValue); err != nil {
		return err
	}

	rv, err := v.readReg(regSPR)

	if err != nil {
		return err
	}

	if rv != testValue {
		return ErrorScratchpad
	}

	return nil
}

var ErrUnsupportedDataBits error = errors.New("unsupported databits")

func (v *SC16IS7X0) setUARTAttributes(dataBits byte, stopBits StopBits, parity Parity) error {

	lcr := byte(0)

	if dataBits == 5 {
		lcr = 0
	} else if dataBits == 6 {
		lcr = 1
	} else if dataBits == 7 {
		lcr = 2
	} else if dataBits == 8 {
		lcr = 3
	} else {
		return ErrUnsupportedDataBits
	}

	if stopBits != Stop1 {
		lcr |= 1 << 2
	}

	if parity == ParityNone {
		//
	} else if parity == ParityOdd {
		lcr |= 1 << 3
	} else if parity == ParityEven {
		lcr |= 3 << 3
	} else if parity == ParityMark {
		lcr |= 5 << 3
	} else if parity == ParitySpace {
		lcr |= 7 << 3
	}

	return v.writeReg(regLCR, lcr)

}

var ErrTimeout error = errors.New("timeout reached")

func (v *SC16IS7X0) WriteByte(c byte) error {
	timeoutTime := time.Now().Add(v.timeout)
	for {
		empty, err := v.peekRegBit(regLSR, 5)
		if err != nil {
			return err
		}
		if empty {
			break
		}
		if timeoutTime.Before(time.Now()) {
			return ErrTimeout
		}
	}

	return v.writeReg(regTHR, c)
}

func (v *SC16IS7X0) Write(p []byte) (n int, err error) {
	written := 0
	for _, val := range p {
		if err := v.WriteByte(val); err != nil {
			return -1, err
		}
		written++
	}

	return written, nil
}

func (v *SC16IS7X0) ReadByte() (byte, error) {
	return v.readReg(regRHR)
}

func (v *SC16IS7X0) Read(p []byte) (int, error) {
	timeoutTime := time.Now().Add(v.timeout)

	for {
		available, err := v.peekRegBit(regLSR, 0)
		if err != nil {
			return -1, err
		}
		if available {
			break
		}
		if timeoutTime.Before(time.Now()) {
			return -2, ErrTimeout
		}
	}

	blen := len(p)
	flen, err := v.fifoDataLength()
	if err != nil {
		return -1, err
	}

	rlen := flen
	if blen < flen {
		rlen = blen
	}

	for i := 0; i < rlen; i++ {
		p[i], err = v.ReadByte()
		if err != nil {
			return i, err
		}
	}

	return rlen, nil
}

type FifoType byte

const (
	RxDFifo FifoType = iota
	TxDFifo
	BothFifo
)

func (v *SC16IS7X0) ClearFifo(clearTarget FifoType) error {
	iirFlags, err := v.readReg(regIIR)
	if err != nil {
		return err
	}

	fcrFlags := byte(0)

	if (iirFlags & 0xC0) != 0 {
		//FIFO enabled
		fcrFlags = 1
	}

	if clearTarget == RxDFifo {
		fcrFlags |= 0x02
	} else if clearTarget == TxDFifo {
		fcrFlags |= 0x04
	} else if clearTarget == BothFifo {
		fcrFlags |= 0x06
	}

	return v.writeReg(regFCR, fcrFlags)
}

func (v *SC16IS7X0) enableFifo(useFifo bool) error {
	if useFifo {
		return v.writeReg(regFCR, 1)
	}
	return v.writeReg(regFCR, 0)
}

func (v *SC16IS7X0) fifoDataLength() (int, error) {
	len, err := v.readReg(regRXLVL)
	if err != nil {
		return -1, err
	}
	return int(len), nil
}

func Open(conf *Config) (*SC16IS7X0, error) {
	size, parityType, stopBits, bus, timeout := conf.Size, conf.Parity, conf.StopBits, conf.Bus, conf.Timeout
	if size == 0 {
		size = DefaultSize
	}
	if parityType == 0 {
		parityType = ParityNone
	}
	if stopBits == 0 {
		stopBits = Stop1
	}
	if bus == 0 {
		bus = 1
	}
	if timeout == 0 {
		timeout = time.Duration(1) * time.Second
	}

	i2c, err := i2c.NewI2C(conf.Address, bus)
	if err != nil {
		return nil, err
	}

	dev := &SC16IS7X0{dev: i2c, xtalFreq: conf.XtalFreq, timeout: timeout}

	if err = dev.testChip(); err != nil {
		dev.Close()
		return nil, err
	}

	if err = dev.setBaudRate(conf.Baud); err != nil {
		dev.Close()
		return nil, err
	}

	if err = dev.enableFifo(true); err != nil {
		dev.Close()
		return nil, err
	}

	if err = dev.setUARTAttributes(size, stopBits, parityType); err != nil {
		dev.Close()
		return nil, err
	}

	if err = dev.ClearFifo(BothFifo); err != nil {
		dev.Close()
		return nil, err
	}

	return dev, nil
}

func (v *SC16IS7X0) Close() error {
	return v.dev.Close()
}
