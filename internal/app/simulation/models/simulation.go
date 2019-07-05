package models

import (
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	pbts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/bburch01/FOTAAS/api"
)

type Simulation struct {
	ID                       string
	DurationInMinutes        int32
	SampleRate               api.SampleRate
	SimulationRateMultiplier api.SimulationRateMultiplier
	GrandPrix                api.GrandPrix
	Track                    api.Track
	State                    string
	StartTimestamp           *pbts.Timestamp
	EndTimestamp             *pbts.Timestamp
	PercentComplete          int32
	FinalStatusCode          string
	FinalStatusMessage       string
	SimulationMembers        map[string]SimulationMember
}

func (sim *Simulation) Create() error {

	sqlStatement := `
			INSERT INTO simulation (id, duration_in_minutes, sample_rate, grand_prix, track,
				 state, start_timestamp, end_timestamp, percent_complete, final_status_code,
				  final_status_message)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GrandPrix.String(), sim.Track.String(),
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

func (sim *Simulation) Update() error {

	sqlStatement := `
			INSERT INTO simulation (id, duration_in_minutes, sample_rate, grand_prix, track,
				 state, start_timestamp, end_timestamp, percent_complete, final_status_code,
				  final_status_message)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	if sim.StartTimestamp == nil && sim.EndTimestamp == nil {
		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GrandPrix.String(), sim.Track.String(),
			sim.State, nil, nil, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
		if err != nil {
			return err
		}
	} else if sim.StartTimestamp == nil && sim.EndTimestamp != nil {
		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GrandPrix.String(), sim.Track.String(),
			sim.State, nil, sim.EndTimestamp, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
		if err != nil {
			return err
		}
	} else if sim.StartTimestamp != nil && sim.EndTimestamp == nil {
		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GrandPrix.String(), sim.Track.String(),
			sim.State, sim.StartTimestamp, nil, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
		if err != nil {
			return err
		}
	} else {
		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate.String(), sim.GrandPrix.String(), sim.Track.String(),
			sim.State, sim.StartTimestamp, sim.EndTimestamp, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
		if err != nil {
			return err
		}
	}

	for _, v := range sim.SimulationMembers {
		err := v.Create()
		if err != nil {
			return err
		}
	}

	return nil
}

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

		/*
			err := rows.Scan(&member.ID, &member.SimulationID, &member.Constructor,
				&member.CarNumber, &member.ForceAlarm, &member.NoAlarms, &member.AlarmOccurred, &member.AlarmDatumDescription,
				&member.AlarmDatumUnit, &member.AlarmDatumValue)
		*/

		err := rows.Scan(&member.ID, &member.SimulationID, &constructor,
			&member.CarNumber, &member.ForceAlarm, &member.NoAlarms, &member.AlarmOccurred, &member.AlarmDatumDescription,
			&member.AlarmDatumUnit, &member.AlarmDatumValue)

		if err != nil {
			return nil, err
		}

		switch constructor {
		case "HAAS":
			member.Constructor = api.Constructor_HAAS
		case "MERCEDES":
			member.Constructor = api.Constructor_MERCEDES
		case "WILLIAMS":
			member.Constructor = api.Constructor_WILLIAMS
		default:
			member.Constructor = api.Constructor_FERRARI
		}

		simMembers = append(simMembers, member)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return simMembers, nil
}

func NewFromRunSimulationRequest(req api.RunSimulationRequest) *Simulation {

	sim := new(Simulation)
	sim.ID = req.Simulation.Uuid
	sim.DurationInMinutes = req.Simulation.DurationInMinutes
	sim.SampleRate = req.Simulation.SampleRate
	sim.SimulationRateMultiplier = req.Simulation.SimulationRateMultiplier
	sim.GrandPrix = req.Simulation.GrandPrix
	sim.Track = req.Simulation.Track

	var simMember SimulationMember
	for _, v := range req.Simulation.SimulationMemberMap {
		simMember.ID = v.Uuid
		simMember.SimulationID = v.SimulationUuid
		simMember.Constructor = v.Constructor
		simMember.CarNumber = v.CarNumber
		simMember.ForceAlarm = v.ForceAlarm
		simMember.NoAlarms = v.NoAlarms
		sim.SimulationMembers[simMember.ID] = simMember
	}

	return sim
}
