package models

import (
	"errors"
	"fmt"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	pbts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/bburch01/FOTAAS/api"
)

type TelemetryDatum struct {
	ID                               string
	Simulated                        bool
	SimulationID                     string
	SimulationTransmitSequenceNumber int32
	GranPrix                         string
	Track                            string
	Constructor                      string
	CarNumber                        int32
	Timestamp                        *pbts.Timestamp
	Latitude                         float64
	Longitude                        float64
	Elevation                        float64
	Description                      string
	Unit                             string
	Value                            float64
	HiAlarm                          bool
	LoAlarm                          bool
}

func (td *TelemetryDatum) Create() error {

	sqlStatement := `
		INSERT INTO telemetry_datum (id, simulated, simulation_id, simulation_transmit_sequence_number, gran_prix, track, constructor,
			car_number, timestamp, latitude, longitude, elevation, description, unit, value, hi_alarm, lo_alarm)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	var t time.Time
	t, err = ipbts.Timestamp(td.Timestamp)
	if err != nil {
		return err
	}

	// Format the timestamp to what mysql likes
	ts := t.Format("2006-01-02 15:04:05")

	_, err = pstmt.Exec(td.ID, td.Simulated, td.SimulationID, td.SimulationTransmitSequenceNumber, td.GranPrix, td.Track, td.Constructor,
		td.CarNumber, ts, td.Latitude, td.Longitude, td.Elevation, td.Description,
		td.Unit, td.Value, td.HiAlarm, td.LoAlarm)
	if err != nil {
		return err
	}

	return nil

}

func RetrieveSimulatedTelemetryData(req api.GetSimulatedTelemetryDataRequest) (api.TelemetryData, error) {

	data := api.TelemetryData{}
	datumMap := make(map[string]*api.TelemetryDatum)

	var txSeqNum, carNumber int32
	var granPrix, track, constructor, datumDescription, datumUnit string
	var ts time.Time

	rows, err := db.Query("select * from telemetry_datum where simulation_id = ? and constructor = ? and car_number = ? and description = ?",
		req.SimulationUuid, req.Constructor.String(), req.CarNumber, req.DatumDescription.String())

	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {

		datum := api.TelemetryDatum{}

		err := rows.Scan(&datum.Uuid, &datum.Simulated, &datum.SimulationUuid, &txSeqNum, &granPrix,
			&track, &constructor, &carNumber, &ts, &datum.Latitude, &datum.Longitude, &datum.Elevation, &datumDescription,
			&datumUnit, &datum.Value, &datum.HighAlarm, &datum.LowAlarm)

		if err != nil {
			return data, err
		}

		ordinal, ok := api.TelemetryDatumDescription_value[datumDescription]
		if !ok {
			return data, fmt.Errorf("invalid telemetry datum description enum: %v", datumDescription)
		}
		datum.Description = api.TelemetryDatumDescription(ordinal)

		ordinal, ok = api.TelemetryDatumUnit_value[datumUnit]
		if !ok {
			return data, fmt.Errorf("invalid telemetry datum unit enum: %v", datumUnit)
		}
		datum.Unit = api.TelemetryDatumUnit(ordinal)

		tsProto, err := ipbts.TimestampProto(ts)
		if err != nil {
			return data, errors.New("failed to convert timestamp to protobuf format")
		}
		datum.Timestamp = tsProto

		//TODO: GranPrix & Track need to be retrieved with a GetSimulation grpc call to the
		//simulation service. This is a hack to just use the values from the final retrieved
		//datum (even if those values *should* always be correct).
		ordinal, ok = api.GranPrix_value[granPrix]
		if !ok {
			return data, fmt.Errorf("invalid telemetry gran prix enum: %v", granPrix)
		}
		data.GranPrix = api.GranPrix(ordinal)

		ordinal, ok = api.Track_value[track]
		if !ok {
			return data, fmt.Errorf("invalid telemetry track enum: %v", track)
		}
		data.Track = api.Track(ordinal)

		datumMap[datum.Uuid] = &datum

	}
	err = rows.Err()
	if err != nil {
		return data, err
	}

	data.Constructor = req.Constructor
	data.CarNumber = req.CarNumber
	data.TelemetryDatumMap = datumMap

	return data, nil

}

func RetrieveTelemetryData(req api.GetTelemetryDataRequest) (api.TelemetryData, error) {

	data := api.TelemetryData{}

	return data, nil

}
