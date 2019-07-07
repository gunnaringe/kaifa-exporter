package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tarm/serial"
	"net/http"
	"time"
)

//import "github.com/howeyc/crc16"

var (
	verbose       = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	listenAddress = kingpin.Flag("listen", "Prometheus listen addr").Short('l').Default(":9500").String()
	serialName    = kingpin.Flag("serial-device", "Device for serial port").Short('d').Default("/dev/ttyUSB0").String()
	serialBaud    = kingpin.Flag("serial-baud", "Baud").Short('b').Default("2400").Int()
	serialParity  = kingpin.Flag("serial-parity", "Parity").Short('p').Default("even").Enum("none", "odd", "even", "mark", "space")
)

type result struct {
	elements  uint8
	timestamp time.Time
	/* Active power (Q1+Q4) */
	activePowerImported uint32
	activePowerExported uint32
	/* OBIL List version identifier */
	obil                           string
	meterId                        string
	meterType                      string
	reactiveImportPower            uint32
	reactiveExportPower            uint32
	currentL1                      float64
	currentL2                      float64
	currentL3                      float64
	voltageL1                      float64
	voltageL2                      float64
	voltageL3                      float64
	meterClock                     time.Time
	cumulativeActiveImportEnergy   uint32
	cumulativeActiveExportEnergy   uint32
	cumulativeReactiveImportEnergy uint32
	cumulativeReactiveExportEnergy uint32
}

func parseTime(b []byte) time.Time {
	year := int(binary.BigEndian.Uint16(b[2:4]))
	month := int(b[4])
	day := int(b[5])
	hour := int(b[7])
	minute := int(b[8])
	second := int(b[9])
	l, _ := time.LoadLocation("Europe/Oslo")
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, l)
}

func readByte(b []byte) (byte, int) {
	if b[0] != 0x02 {
		panic("Type is not a byte")
	}
	return b[1], 2
}

func readInt32(b []byte) (uint32, int) {
	if b[0] != 0x06 {
		panic(fmt.Sprintf("Type is not a uint32: %0X", b[0]))
	}
	return binary.BigEndian.Uint32(b[1:]), 5
}

func readInt16(b []byte) (uint16, int) {
	if b[0] != 0x12 {
		panic(fmt.Sprintf("Type is not a uint16: %0X", b[0]))
	}
	return binary.BigEndian.Uint16(b[1:]), 3
}

func readString(b []byte) (string, int) {
	if b[0] != 0x09 {
		panic(fmt.Sprintf("Type is not a uint16: %0X", b[0]))
	}
	length := int(b[1])
	return string(b[2 : length+2]), length + 2
}

func handle(data []byte) {
	var activePowerImported uint32
	var activePowerExported uint32
	var reactiveImportPower uint32
	var reactiveExportPower uint32
	var currentL1 uint32
	var currentL2 uint32
	var currentL3 uint32
	var voltageL1 uint32
	var voltageL2 uint32
	var voltageL3 uint32
	var obil string
	var meterId string
	var meterType string
	var meterClock time.Time
	var cumulativeActiveImportEnergy uint32
	var cumulativeActiveExportEnergy uint32
	var cumulativeReactiveImportEnergy uint32
	var cumulativeReactiveExportEnergy uint32

	obisCodeValue := data[0]
	length := data[1]
	//source := data[2:4]
	//destination := data[4]
	//controlField := data[5] // crc16 checksum
	//hsc := data[6:8]
	if *verbose {
		fmt.Printf("%0X obis code value\n", obisCodeValue)
		fmt.Printf("%d bytes", length)
	}
	//fmt.Printf("DLMS/COSEM LLC Addresses: %X\n", data[8:11])
	//fmt.Printf("DLMS HEADER?: %X\n", data[11:16])

	// Skipping headers for now
	pos := 30
	elements, bytesConsumed := readByte(data[pos:])
	pos += bytesConsumed

	// List 1
	if elements == 1 {
		activePowerImported, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed
	}

	// List 2 + List 3
	if elements == 13 || elements == 18 {
		obil, bytesConsumed = readString(data[pos:])
		pos += bytesConsumed

		meterId, bytesConsumed = readString(data[pos:])
		pos += bytesConsumed

		meterType, bytesConsumed = readString(data[pos:])
		pos += bytesConsumed

		activePowerImported, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		activePowerExported, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		reactiveImportPower, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		reactiveExportPower, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		currentL1, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		currentL2, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		currentL3, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		voltageL1, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		voltageL2, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		voltageL3, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed
	}

	// List 3
	if elements == 18 {
		meterClockSize := 14
		meterClock = parseTime(data[pos : pos+meterClockSize])
		pos += meterClockSize

		cumulativeActiveImportEnergy, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		cumulativeActiveExportEnergy, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		cumulativeReactiveImportEnergy, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		cumulativeReactiveExportEnergy, bytesConsumed = readInt32(data[pos:])
		pos += bytesConsumed

		//fmt.Printf("Remaining: % X\n", data[pos:])
	}

	list := result{
		elements:                       elements,
		timestamp:                      parseTime(data[16:30]),
		activePowerImported:            activePowerImported,
		activePowerExported:            activePowerExported,
		reactiveImportPower:            reactiveImportPower,
		reactiveExportPower:            reactiveExportPower,
		currentL1:                      float64(currentL1) / 1000.0,
		currentL2:                      float64(currentL2) / 1000.0,
		currentL3:                      float64(currentL3) / 1000.0,
		voltageL1:                      float64(voltageL1) / 10.0,
		voltageL2:                      float64(voltageL2) / 10.0,
		voltageL3:                      float64(voltageL3) / 10.0,
		obil:                           obil,
		meterId:                        meterId,
		meterType:                      meterType,
		meterClock:                     meterClock,
		cumulativeActiveImportEnergy:   cumulativeActiveImportEnergy,
		cumulativeActiveExportEnergy:   cumulativeActiveExportEnergy,
		cumulativeReactiveImportEnergy: cumulativeReactiveImportEnergy,
		cumulativeReactiveExportEnergy: cumulativeReactiveExportEnergy,
	}

	register(&list)
	if *verbose {
		write(&list)
	}
}

func write(list *result) {
	fmt.Printf("\n----------------------------------------------------------------------------\n")
	if list.elements == 1 {
		fmt.Printf("           active: %6d W imported\n", list.activePowerImported)
	}
	if list.elements == 13 || list.elements == 18 {
		fmt.Printf("        timestamp: %s\n", list.timestamp)

		fmt.Println()
		fmt.Printf("                         L1       L2       L3\n")
		fmt.Printf("          voltage: %6.1f V %6.1f V %6.1f V\n", list.voltageL1, list.voltageL2, list.voltageL3)
		fmt.Printf("          current: %6.1f A %6.1f A %6.1f A\n", list.currentL1, list.currentL2, list.currentL3)

		fmt.Println()
		fmt.Printf("                       IMPORT         EXPORT\n")
		fmt.Printf("           active: %8d W     %8d W\n", list.activePowerImported, list.activePowerExported)
		fmt.Printf("         reactive: %8d VAr   %8d VAr\n", list.reactiveImportPower, list.reactiveExportPower)
	}
	if list.elements == 18 {
		fmt.Printf("    active energy: %8d WH    %8d WH\n", list.cumulativeActiveImportEnergy, list.cumulativeActiveExportEnergy)
		fmt.Printf("  reactive energy: %8d VArh  %8d VArh\n", list.cumulativeReactiveImportEnergy, list.cumulativeActiveExportEnergy)
	}
}

func parity(name *string) serial.Parity {
	switch *name {
	case "none":
		return serial.ParityNone
	case "odd":
		return serial.ParityOdd
	case "even":
		return serial.ParityEven
	case "mark":
		return serial.ParityMark
	case "space":
		return serial.ParitySpace
	default:
		panic("Invalid parity")
	}
}

func main() {
	kingpin.Parse()
	fmt.Println("Starting read of HAN port")

	parity := parity(serialParity)
	c := &serial.Config{Name: *serialName, Baud: *serialBaud, Parity: parity, StopBits: serial.Stop1}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(s)

	// Throw away until we find first delimiter
	_, err = reader.ReadBytes('\x7E')
	if err != nil {
		panic(err)
	}

	go func() {
		for true {
			// Read until hitting delimiter
			bytes, err := reader.ReadBytes('\x7E')
			if err != nil {
				panic(err)
			}

			// Start delimiter will result in a single item - Throw
			readLength := len(bytes)
			if readLength <= 2 {
				continue
			}

			// Read length includes stop bit
			declaredLength := int(bytes[1])
			if declaredLength != readLength-1 {
				fmt.Printf("[Skipping] Read and declared message length does not match: actual=%d declared:%d\n", readLength-1, declaredLength)
				continue
			}

			if *verbose {
				fmt.Printf("Length: %d\n", len(bytes))
			}
			handle(bytes)
		}
	}()

	fmt.Printf("Serving metrics at %s\n", *listenAddress)
	http.Handle("/", promhttp.Handler())
	fmt.Println(http.ListenAndServe(*listenAddress, nil))
}
