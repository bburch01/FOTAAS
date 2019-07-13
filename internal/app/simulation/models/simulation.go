package models

import (
	"errors"
	"fmt"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	itime "github.com/bburch01/FOTAAS/internal/pkg/time"
	pbts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/bburch01/FOTAAS/api"
)

type Simulation struct {
	ID                       string
	DurationInMinutes        int32
	SampleRate               api.SampleRate
	SimulationRateMultiplier api.SimulationRateMultiplier
	GranPrix                 api.GranPrix
	Track                    api.Track
	State                    string
	StartTimestamp           *pbts.Timestamp
	EndTimestamp             *pbts.Timestamp
	PercentComplete          float32
	FinalStatusCode          string
	FinalStatusMessage       string
	SimulationMembers        map[string]SimulationMember
}

func (sim *Simulation) Create() error {

	sqlStatement := `
			INSERT INTO simulation (id, duration_in_minutes, sample_rate, gran_prix, track,
				 state, start_timestamp, end_timestamp, percent_complete, final_status_code,
				  final_status_message)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GranPrix.String(), sim.Track.String(),
		sim.State, nil, nil, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
	if err != nil {
		return err
	}

	for _, v := range sim.SimulationMembers {
		err := v.Create()
		if err != nil {
			return err
		}
	}

	return nil
}

/*
func (sim *Simulation) Retrieve() error {

	var sampleRate, granPrix, track, state string
	var startTs, endTs time.Time

	rows, err := db.Query("select * from simulation where id = ?", sim.ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&sim.ID, &sim.DurationInMinutes,
			&sampleRate, &granPrix, &track, &state, &startTs, &endTs, &sim.PercentComplete,
			&sim.FinalStatusCode, &sim.FinalStatusMessage)

		if err != nil {
			return nil, err
		}

		ordinal, ok := api.SampleRate_value[sampleRate]
		if !ok {
			return nil, fmt.Errorf("invalid simulation sample rate enum: %v", sampleRate)
		}
		sim.SampleRate = api.SampleRate(ordinal)

		ordinal, ok = api.GranPrix_value[granPrix]
		if !ok {
			return data, fmt.Errorf("invalid simulation gran prix enum: %v", granPrix)
		}
		sim.GranPrix = api.GranPrix(ordinal)

		ordinal, ok = api.Track_value[track]
		if !ok {
			return data, fmt.Errorf("invalid simulation track enum: %v", track)
		}
		sim.Track = api.Track(ordinal)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

}
*/

func (sim Simulation) UpdateState() error {

	sqlStatement := `UPDATE simulation SET state = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.State, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) UpdateStartTimestamp() error {

	if sim.StartTimestamp == nil {
		return errors.New("simulation StartTimestamp must not be nil")
	}

	var t time.Time

	sqlStatement := `UPDATE simulation SET start_timestamp = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	// Re-format the timestamp to mysql format
	t, err = ipbts.Timestamp(sim.StartTimestamp)
	if err != nil {
		return err
	}
	startTs := t.Format("2006-01-02 15:04:05")

	_, err = pstmt.Exec(startTs, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) UpdateEndTimestamp() error {

	if sim.EndTimestamp == nil {
		return errors.New("simulation EndTimestamp must not be nil")
	}

	var t time.Time

	sqlStatement := `UPDATE simulation SET end_timestamp = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	// Re-format the timestamp to mysql format
	t, err = ipbts.Timestamp(sim.EndTimestamp)
	if err != nil {
		return err
	}
	endTs := t.Format("2006-01-02 15:04:05")

	_, err = pstmt.Exec(endTs, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) UpdatePercentComplete() error {

	sqlStatement := `UPDATE simulation SET percent_complete = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.PercentComplete, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) UpdateFinalStatusCode() error {

	sqlStatement := `UPDATE simulation SET final_status_code = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.FinalStatusCode, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) UpdateFinalStatusMessage() error {

	sqlStatement := `UPDATE simulation SET final_status_message = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.FinalStatusMessage, sim.ID)
	if err != nil {
		return err
	}

	return nil
}

func (sim Simulation) FindAllMembers() ([]SimulationMember, error) {

	var simMembers []SimulationMember
	var member SimulationMember
	var constructor string

	rows, err := db.Query("select * from simulation_member where simulation_id = ?", sim.ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&member.ID, &member.SimulationID, &constructor,
			&member.CarNumber, &member.ForceAlarm, &member.NoAlarms, &member.AlarmOccurred, &member.AlarmDatumDescription,
			&member.AlarmDatumUnit, &member.AlarmDatumValue)

		if err != nil {
			return nil, err
		}

		ordinal, ok := api.Constructor_value[constructor]
		if !ok {
			return nil, fmt.Errorf("invalid constructor enum: %v", constructor)
		}
		member.Constructor = api.Constructor(ordinal)
		simMembers = append(simMembers, member)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return simMembers, nil
}

func RetrieveSimulationInfo(req api.GetSimulationInfoRequest) (*api.SimulationInfo, error) {

	var sampleRate, granPrix, track, state string
	var startTs, endTs itime.NullTime

	info := api.SimulationInfo{}

	rows, err := db.Query("select * from simulation where id = ?", req.SimulationUuid)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&info.Uuid, &info.DurationInMinutes,
			&sampleRate, &granPrix, &track, &state, &startTs, &endTs, &info.PercentComplete,
			&info.FinalStatusCode, &info.FinalStatusMessage)

		if err != nil {
			return nil, err
		}

		ordinal, ok := api.SampleRate_value[sampleRate]
		if !ok {
			return nil, fmt.Errorf("invalid simulation sample rate enum: %v", sampleRate)
		}
		info.SampleRate = api.SampleRate(ordinal)

		ordinal, ok = api.GranPrix_value[granPrix]
		if !ok {
			return nil, fmt.Errorf("invalid simulation gran prix enum: %v", granPrix)
		}
		info.GranPrix = api.GranPrix(ordinal)

		ordinal, ok = api.Track_value[track]
		if !ok {
			return nil, fmt.Errorf("invalid simulation track enum: %v", track)
		}
		info.Track = api.Track(ordinal)

		ordinal, ok = api.SimulationState_value[state]
		if !ok {
			return nil, fmt.Errorf("invalid simulation state enum: %v", track)
		}
		info.State = api.SimulationState(ordinal)

		tsProto, err := ipbts.TimestampProto(startTs.Time)
		if err != nil {
			return nil, errors.New("failed to convert start timestamp to protobuf format")
		}
		info.StartTimestamp = tsProto

		tsProto, err = ipbts.TimestampProto(endTs.Time)
		if err != nil {
			return nil, errors.New("failed to convert end timestamp to protobuf format")
		}
		info.EndTimestamp = tsProto

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &info, nil

}

func NewFromRunSimulationRequest(req api.RunSimulationRequest) *Simulation {

	sim := new(Simulation)
	sim.SimulationMembers = make(map[string]SimulationMember)
	sim.ID = req.Simulation.Uuid
	sim.DurationInMinutes = req.Simulation.DurationInMinutes
	sim.SampleRate = req.Simulation.SampleRate
	sim.SimulationRateMultiplier = req.Simulation.SimulationRateMultiplier
	sim.GranPrix = req.Simulation.GranPrix
	sim.Track = req.Simulation.Track

	var simMember SimulationMember
	for _, v := range req.Simulation.SimulationMemberMap {
		simMember.ID = v.Uuid
		simMember.SimulationID = v.SimulationUuid
		simMember.Constructor = v.Constructor
		simMember.CarNumber = v.CarNumber
		simMember.ForceAlarm = v.ForceAlarm
		simMember.NoAlarms = v.NoAlarms
		sim.SimulationMembers[v.Uuid] = simMember
	}

	return sim
}

/*
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
*/
