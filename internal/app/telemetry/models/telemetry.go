package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func RetrieveTelemetryData(req api.GetTelemetryDataRequest) (*api.TelemetryData, error) {

	data := api.TelemetryData{}
	datumMap := make(map[string]*api.TelemetryDatum)

	var txSeqNum, carNumber int32
	var granPrix, track, constructor, datumDescription, datumUnit string
	var ts time.Time

	// Build the select query based on the search by flags in the request. If none of the search flags
	// are set, select by simulation uuid only.
	var sb strings.Builder

	if req.Simulated || req.SimulationUuid != "" {
		if req.SimulationUuid != "" {
			sb.WriteString("select * from telemetry_datum where simulation_id = '")
			sb.WriteString(req.SimulationUuid)
			sb.WriteString("'")
		} else {
			sb.WriteString("select * from telemetry_datum where simulated = true")
		}
	} else {
		sb.WriteString("select * from telemetry_datum where simulated = false")
	}

	if req.SearchBy.Constructor {
		sb.WriteString(" and constructor = '")
		sb.WriteString(req.Constructor.String())
		sb.WriteString("'")
	}
	if req.SearchBy.CarNumber {
		sb.WriteString(" and car_number = ")
		sb.WriteString(strconv.Itoa(int(req.CarNumber)))
	}
	if req.SearchBy.DatumDescription {
		sb.WriteString(" and description = '")
		sb.WriteString(req.DatumDescription.String())
		sb.WriteString("'")
	}
	if req.SearchBy.GranPrix {
		sb.WriteString(" and gran_prix = '")
		sb.WriteString(req.GranPrix.String())
		sb.WriteString("'")
	}
	if req.SearchBy.Track {
		sb.WriteString(" and track = '")
		sb.WriteString(req.Track.String())
		sb.WriteString("'")
	}
	if req.SearchBy.HighAlarm && !req.SearchBy.LowAlarm {
		sb.WriteString(" and hi_alarm = true")
	}
	if req.SearchBy.LowAlarm && !req.SearchBy.HighAlarm {
		sb.WriteString(" and lo_alarm = true")
	}
	if req.SearchBy.LowAlarm && req.SearchBy.HighAlarm {
		sb.WriteString(" and (lo_alarm = true or hi_alarm = true)")
	}
	if req.SearchBy.DateRange {

		var startTs, endTs time.Time
		var err error

		if startTs, err = ipbts.Timestamp(req.DateRangeBegin); err != nil {
			return nil, err
		}

		if endTs, err = ipbts.Timestamp(req.DateRangeEnd); err != nil {
			return nil, err
		}

		sb.WriteString(" and timestamp between '")

		sb.WriteString(strconv.Itoa(startTs.Year()))
		sb.WriteString("-")
		sb.WriteString(strconv.Itoa(int(startTs.Month())))
		sb.WriteString("-")
		sb.WriteString(strconv.Itoa(startTs.Day()))
		sb.WriteString(" 00:00:00")
		sb.WriteString("' and '")
		sb.WriteString(strconv.Itoa(endTs.Year()))
		sb.WriteString("-")
		sb.WriteString(strconv.Itoa(int(endTs.Month())))
		sb.WriteString("-")
		sb.WriteString(strconv.Itoa(endTs.Day()))
		sb.WriteString(" 23:59:59")
		sb.WriteString("'")
	}

	logger.Debug(fmt.Sprintf("select sql: %v", sb.String()))
	rows, err := db.Query(sb.String())

	switch {
	case err == sql.ErrNoRows:
		// no rows & no errors, caller needs to check for nil TelemetryData
		return nil, nil
	case err != nil:
		logger.Error(fmt.Sprintf("failed to retrieve telemetry data with error: %v", err))
		return nil, err
	default:

		defer rows.Close()

		for rows.Next() {

			datum := api.TelemetryDatum{}

			err := rows.Scan(&datum.Uuid, &datum.Simulated, &datum.SimulationUuid, &txSeqNum, &granPrix,
				&track, &constructor, &carNumber, &ts, &datum.Latitude, &datum.Longitude, &datum.Elevation, &datumDescription,
				&datumUnit, &datum.Value, &datum.HighAlarm, &datum.LowAlarm)

			if err != nil {
				return nil, err
			}

			ordinal, ok := api.TelemetryDatumDescription_value[datumDescription]
			if !ok {
				return nil, fmt.Errorf("invalid telemetry datum description enum: %v", datumDescription)
			}
			datum.Description = api.TelemetryDatumDescription(ordinal)

			ordinal, ok = api.TelemetryDatumUnit_value[datumUnit]
			if !ok {
				return nil, fmt.Errorf("invalid telemetry datum unit enum: %v", datumUnit)
			}
			datum.Unit = api.TelemetryDatumUnit(ordinal)

			tsProto, err := ipbts.TimestampProto(ts)
			if err != nil {
				return nil, errors.New("failed to convert timestamp to protobuf format")
			}
			datum.Timestamp = tsProto

			ordinal, ok = api.GranPrix_value[granPrix]
			if !ok {
				return nil, fmt.Errorf("invalid gran prix enum: %v", granPrix)
			}
			datum.GranPrix = api.GranPrix(ordinal)

			ordinal, ok = api.Track_value[track]
			if !ok {
				return nil, fmt.Errorf("invalid track enum: %v", track)
			}
			datum.Track = api.Track(ordinal)

			ordinal, ok = api.Constructor_value[constructor]
			if !ok {
				return nil, fmt.Errorf("invalid constructor enum: %v", constructor)
			}
			datum.Constructor = api.Constructor(ordinal)

			datum.CarNumber = carNumber

			datumMap[datum.Uuid] = &datum

		}
		err = rows.Err()
		if err != nil {
			return nil, err
		}

		data.TelemetryDatumMap = datumMap

	}

	return &data, nil

}
