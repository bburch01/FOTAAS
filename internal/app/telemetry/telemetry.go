package telemetry

import (
	"github.com/bburch01/FOTAAS/api"
)

type AlarmMode int

const (
	High AlarmMode = iota
	Low
)

func (am AlarmMode) String() string {
	return [...]string{"High", "Low"}[am]
}

type TelemetryDatumParameters struct {
	Unit           api.TelemetryDatumUnit
	RangeLowValue  float64
	RangeHighValue float64
	HighAlarmValue float64
	LowAlarmValue  float64
}

type SimulatedTelemetryData struct {
	DatumDesc   api.TelemetryDatumDescription
	Data        []api.TelemetryDatum
	AlarmExists bool
	AlarmMode   AlarmMode
	AlarmIndex  int
}

type AlarmParams struct {
	Desc api.TelemetryDatumDescription
	Mode AlarmMode
}
