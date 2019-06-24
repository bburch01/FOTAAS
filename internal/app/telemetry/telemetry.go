package telemetry

import (
	pb "github.com/bburch01/FOTAAS/api"
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
	Unit           pb.TelemetryDatumUnit
	RangeLowValue  float64
	RangeHighValue float64
	HighAlarmValue float64
	LowAlarmValue  float64
}

type SimulatedTelemetryData struct {
	DatumDesc   pb.TelemetryDatumDescription
	Data        []pb.TelemetryDatum
	AlarmExists bool
	AlarmMode   AlarmMode
	AlarmIndex  int
}

type AlarmParams struct {
	Desc pb.TelemetryDatumDescription
	Mode AlarmMode
}
