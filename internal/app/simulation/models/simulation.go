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
	SampleRate               string
	SimulationRateMultiplier string
	GrandPrix                string
	Track                    string
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

	_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate, sim.GrandPrix, sim.Track,
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

	_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate, sim.GrandPrix, sim.Track,
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

	rows, err := db.Query("select * from simulation_member where simulation_id = ?", sim.ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&member.ID, &member.SimulationID, &member.Constructor,
			&member.CarNumber, &member.ForceAlarm, &member.NoAlarms, &member.AlarmOccurred, &member.AlarmDatumDescription,
			&member.AlarmDatumUnit, &member.AlarmDatumValue)
		if err != nil {
			return nil, err
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
	sim.SampleRate = req.Simulation.SampleRate.String()
	sim.SimulationRateMultiplier = req.Simulation.SimulationRateMultiplier.String()
	sim.GrandPrix = req.Simulation.GrandPrix.String()
	sim.Track = req.Simulation.Track.String()

	var simMember SimulationMember
	for _, v := range req.Simulation.SimulationMemberMap {
		simMember.ID = v.Uuid
		simMember.SimulationID = v.SimulationUuid
		simMember.Constructor = v.Constructor.String()
		simMember.CarNumber = v.CarNumber
		simMember.ForceAlarm = v.ForceAlarm
		simMember.NoAlarms = v.NoAlarms
		sim.SimulationMembers[simMember.ID] = simMember
	}

	return sim
}
