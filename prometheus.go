package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespace = "kaifa"

	activePowerImported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_power_imported_watts",
		Help:      "Active power in import direction",
	})
	activePowerExported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_power_exported_watts",
		Help:      "Active power in export direction",
	})
	reactivePowerImported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "reactive_power_imported_var",
		Help:      "Reactive power in import direction",
	})
	reactivePowerExported = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "reactive_power_exported_var",
		Help:      "Reactive power in export direction",
	})
	instantCurrent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "current",
			Help:      "Instantaneous current",
		},
		[]string{
			"phase",
		})
	instantCurrentL1 = instantCurrent.WithLabelValues("L1")
	instantCurrentL2 = instantCurrent.WithLabelValues("L2")
	instantCurrentL3 = instantCurrent.WithLabelValues("L3")

	instantVoltage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "voltage",
			Help:      "Instantaneous voltage (phase voltage 4W meter, line voltage 3W meter, 1 second sampling)",
		},
		[]string{
			"phase",
		})
	instantVoltageL1L2 = instantVoltage.WithLabelValues("L1-L2")
	instantVoltageL1L3 = instantVoltage.WithLabelValues("L1-L3")
	instantVoltageL2L3 = instantVoltage.WithLabelValues("L2-L3")
	information        = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Meter information",
		},
		[]string{
			"meter_id",
			"meter_type",
			"obil",
		})
	lastUpdate = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_update_timestamp",
			Help:      "Last update from meter in seconds since Unix epoch",
		})
	timestampHeader = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "timestamp",
			Help:      "Timestamp from meter header in seconds since Unix epoch",
		})
	timestampMeter = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "meter_timestamp",
			Help:      "Meter timestamp in seconds since Unix epoch",
		})
	cumulativeActiveImportEnergy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cumulative_active_import_energy_watt_hours",
			Help:      "Cumulative active import energy (A+)",
		})
	cumulativeActiveExportEnergy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cumulative_active_export_energy_watt_hours",
			Help:      "Cumulative active export energy (A-)",
		})
	cumulativeReactiveImportEnergy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cumulative_reactive_import_energy_var_hours",
			Help:      "Cumulative reactive import energy (R+)",
		})
	cumulativeReactiveExportEnergy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cumulative_reactive_export_energy",
			Help:      "Cumulative reactive export energy (R-)",
		})
)

func register(list *result) {
	elements := list.elements

	activePowerImported.Set(float64(list.activePowerImported))
	timestampHeader.Set(float64(list.timestamp.UnixNano()) / 1e9)
	timestampMeter.Set(float64(list.meterClock.UnixNano()) / 1e9)

	if elements == 13 || elements == 18 {
		information.WithLabelValues(list.meterId, list.meterType, list.obil).Set(1)
		activePowerExported.Set(float64(list.activePowerExported))
		reactivePowerImported.Set(float64(list.reactiveImportPower))
		reactivePowerExported.Set(float64(list.reactiveExportPower))
		instantCurrentL1.Set(list.currentL1)
		instantCurrentL2.Set(list.currentL2)
		instantCurrentL3.Set(list.currentL3)
		instantVoltageL1L2.Set(list.voltageL1)
		instantVoltageL1L3.Set(list.voltageL2)
		instantVoltageL2L3.Set(list.voltageL3)
	}
	if elements == 18 {
		cumulativeActiveImportEnergy.Set(float64(list.cumulativeActiveImportEnergy))
		cumulativeActiveExportEnergy.Set(float64(list.cumulativeActiveExportEnergy))
		cumulativeReactiveImportEnergy.Set(float64(list.cumulativeReactiveImportEnergy))
		cumulativeReactiveExportEnergy.Set(float64(list.cumulativeReactiveExportEnergy))
	}

	lastUpdate.SetToCurrentTime()
}
