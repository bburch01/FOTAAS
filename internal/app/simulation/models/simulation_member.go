package models

import (
	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/telemetry"
)

type SimulationMember struct {
	ID                    string
	SimulationID          string
	Constructor           string
	CarNumber             int32
	TelemetryData         map[api.TelemetryDatumDescription]telemetry.SimulatedTelemetryData
	ForceAlarm            bool
	NoAlarms              bool
	AlarmOccurred         bool
	AlarmDatumDescription string
	AlarmDatumUnit        string
	AlarmDatumValue       float64
}

func (simMember SimulationMember) Create() error {

	sqlStatement := `
		INSERT INTO simulation_member (id, simulation_id, constructor, car_number, force_alarm, no_alarms)
		VALUES (?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(simMember.ID, simMember.SimulationID, simMember.Constructor,
		simMember.CarNumber, simMember.ForceAlarm, simMember.NoAlarms)
	if err != nil {
		return err
	}

	return nil
}

func (simMember SimulationMember) UpdateAlarmInfo() error {

	sqlStatement := `UPDATE simulation_member SET alarm_occurred = ?, alarm_datum_description = ?,
	alarm_datum_unit = ?, alarm_datum_value =? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(simMember.AlarmOccurred, simMember.AlarmDatumDescription,
		simMember.AlarmDatumUnit, simMember.AlarmDatumValue, simMember.ID)
	if err != nil {
		return err
	}

	return nil

}
