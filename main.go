package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tarm/serial"
	"net/http"
	"time"
)

//import "github.com/howeyc/crc16"

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

var (
	prom_activePowerImported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "active_power_imported",
		Help:      "-",
	})
	prom_activePowerExported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "active_power_exported",
		Help:      "-",
	})
	prom_reactivePowerImported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "reactive_power_imported",
		Help:      "-",
	})
	prom_reactivePowerExported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "reactive_power_exported",
		Help:      "-",
	})
	prom_current = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "current",
		Help:      "-",
	},
		[]string{"phase"})
	prom_voltage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "voltage",
		Help:      "-",
	},
		[]string{"phase"})
	prom_info = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "info",
		Help:      "-",
	},
		[]string{"meter_id", "meter_type", "obil"})
	prom_fetch = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "last_update",
		Help:      "-",
	})
	prom_timestamp = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "timestamp",
		Help:      "-",
	})
	prom_metertime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "metertime",
		Help:      "-",
	})
	prom_cumulativeActiveImportEnergy = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "cumulative_active_import_energy",
		Help:      "-",
	})
	prom_cumulativeActiveExportEnergy = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "cumulative_active_export_energy",
		Help:      "-",
	})
	prom_cumulativeReactiveImportEnergy = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "cumulative_reactive_import_energy",
		Help:      "-",
	})
	prom_cumulativeReactiveExportEnergy = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kaifa",
		Name:      "cumulative_reactive_export_energy",
		Help:      "-",
	})
)

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

	// obisCodeValue := data[0]
	// dataLength := data[1]
	// source := data[2:4]
	// destination := data[4]
	// controlField := data[5] // crc16 checksum
	// hsc := data[6:8]
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

	fmt.Printf("\n----------------------------------------------------------------------------\n")
	if elements == 1 {
		fmt.Printf("           active: %6d W imported\n", list.activePowerImported)
	}
	if elements == 13 || elements == 18 {
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
	if elements == 18 {
		fmt.Printf("    active energy: %8d WH    %8d WH\n", list.cumulativeActiveImportEnergy, list.cumulativeActiveExportEnergy)
		fmt.Printf("  reactive energy: %8d VArh  %8d VArh\n", list.cumulativeReactiveImportEnergy, list.cumulativeActiveExportEnergy)
	}

	prom_activePowerImported.Set(float64(list.activePowerImported))
	prom_timestamp.Set(float64(list.timestamp.UnixNano()) / 1e9)
	prom_metertime.Set(float64(list.meterClock.UnixNano()) / 1e9)

	if elements == 13 || elements == 18 {
		prom_info.WithLabelValues(list.meterId, list.meterType, list.obil).Set(1)

		prom_activePowerExported.Set(float64(list.activePowerExported))
		prom_reactivePowerImported.Set(float64(list.reactiveImportPower))
		prom_reactivePowerExported.Set(float64(list.reactiveExportPower))

		prom_current.WithLabelValues("1").Set(list.currentL1)
		prom_current.WithLabelValues("2").Set(list.currentL2)
		prom_current.WithLabelValues("3").Set(list.currentL3)
		prom_voltage.WithLabelValues("1").Set(list.voltageL1)
		prom_voltage.WithLabelValues("2").Set(list.voltageL2)
		prom_voltage.WithLabelValues("3").Set(list.voltageL3)
	}
	if elements == 18 {
		prom_cumulativeActiveImportEnergy.Set(float64(list.cumulativeActiveImportEnergy))
		prom_cumulativeActiveExportEnergy.Set(float64(list.cumulativeActiveExportEnergy))
		prom_cumulativeReactiveImportEnergy.Set(float64(list.cumulativeReactiveImportEnergy))
		prom_cumulativeReactiveExportEnergy.Set(float64(list.cumulativeReactiveExportEnergy))
	}

	prom_fetch.SetToCurrentTime()
}

func main() {
	reply := []byte{'\xA0', '\x9B', '\x01', '\x02', '\x01', '\x10', '\xEE', '\xAE', '\xE6', '\xE7', '\x00', '\x0F', '\x40', '\x00', '\x00', '\x00', '\x09', '\x0C', '\x07', '\xE3', '\x03', '\x09', '\x06', '\x0D', '\x00', '\x0A', '\xFF', '\x80', '\x00', '\x00', '\x02', '\x12', '\x09', '\x07', '\x4B', '\x46', '\x4D', '\x5F', '\x30', '\x30', '\x31', '\x09', '\x10', '\x36', '\x39', '\x37', '\x30', '\x36', '\x33', '\x31', '\x34', '\x30', '\x33', '\x37', '\x39', '\x36', '\x35', '\x35', '\x33', '\x09', '\x08', '\x4D', '\x41', '\x33', '\x30', '\x34', '\x48', '\x33', '\x45', '\x06', '\x00', '\x00', '\x0F', '\xE5', '\x06', '\x00', '\x00', '\x00', '\x00', '\x06', '\x00', '\x00', '\x02', '\xBD', '\x06', '\x00', '\x00', '\x00', '\x00', '\x06', '\x00', '\x00', '\x36', '\x7F', '\x06', '\x00', '\x00', '\x37', '\x90', '\x06', '\x00', '\x00', '\x11', '\x6B', '\x06', '\x00', '\x00', '\x09', '\x36', '\x06', '\x00', '\x00', '\x00', '\x00', '\x06', '\x00', '\x00', '\x09', '\x30', '\x09', '\x0C', '\x07', '\xE3', '\x03', '\x09', '\x06', '\x0D', '\x00', '\x0A', '\xFF', '\x80', '\x00', '\x00', '\x06', '\x01', '\x28', '\x4D', '\x75', '\x06', '\x00', '\x00', '\x00', '\x00', '\x06', '\x00', '\x0E', '\xF7', '\x72', '\x06', '\x00', '\x03', '\x15', '\x35', '\xC3', '\xAC', '\x7E'}

	go func() {
		handle(reply)
	}()

	fmt.Println("Serving metrics at http://localhost:9500/metrics")
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(http.ListenAndServe(":9500", nil))
}

func main2() {
	fmt.Println("Starting read of HAN port")

	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 2400, Parity: serial.ParityEven, StopBits: serial.Stop1}
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

	i := 1
	for i < 2 {
		// Read until hitting delimiter
		reply, err := reader.ReadBytes('\x7E')
		if err != nil {
			panic(err)
		}
		// Start delimiter will result in a single item - Throw
		length := len(reply)
		if length == 1 {
			continue
		}

		fmt.Printf("Length: %d\n", len(reply))
		handle(reply)
	}
}
