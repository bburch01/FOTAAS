package models

import (
	"time"

	pbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type TelemetryDatum struct {
	ID                               string
	Simulated                        bool
	SimulationID                     string
	SimulationTransmitSequenceNumber int32
	GrandPrix                        string
	Track                            string
	Constructor                      string
	CarNumber                        int32
	Timestamp                        *timestamp.Timestamp
	Latitude                         float64
	Longitude                        float64
	Elevation                        float64
	Description                      string
	Unit                             string
	Value                            float64
	HiAlarm                          bool
	LoAlarm                          bool
}

func (td *TelemetryDatum) Persist() error {

	sqlStatement := `
		INSERT INTO telemetry_datum (id, simulated, simulation_id, simulation_transmit_sequence_number, grand_prix, track, constructor,
			car_number, timestamp, latitude, longitude, elevation, description, unit, value, hi_alarm, lo_alarm)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	var t time.Time
	t, err = pbts.Timestamp(td.Timestamp)
	if err != nil {
		return err
	}

	// Format the timestamp to what mysql likes
	ts := t.Format("2006-01-02 15:04:05")

	_, err = pstmt.Exec(td.ID, td.Simulated, td.SimulationID, td.SimulationTransmitSequenceNumber, td.GrandPrix, td.Track, td.Constructor,
		td.CarNumber, ts, td.Latitude, td.Longitude, td.Elevation, td.Description,
		td.Unit, td.Value, td.HiAlarm, td.LoAlarm)
	if err != nil {
		return err
	}

	return nil

}
