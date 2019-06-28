package models

import (
	"time"

	pbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type Simulation struct {
	ID                 string
	DurationInMinutes  int32
	SampleRate         string
	GrandPrix          string
	Track              string
	State              string
	StartTimestamp     *timestamp.Timestamp
	EndTimestamp       *timestamp.Timestamp
	PercentComplete    int32
	FinalStatusCode    string
	FinalStatusMessage string
}

func (sim *Simulation) Create() error {

	var t time.Time

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

	// Re-format the timestamp to mysql format
	t, err = pbts.Timestamp(sim.StartTimestamp)
	if err != nil {
		return err
	}
	startTs := t.Format("2006-01-02 15:04:05")

	// There must be a better way to insert a null timestamp into the mysql db
	if sim.EndTimestamp != nil {
		// Re-format the timestamp to mysql format
		t, err = pbts.Timestamp(sim.EndTimestamp)
		if err != nil {
			return err
		}
		endTs := t.Format("2006-01-02 15:04:05")

		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate, sim.GrandPrix, sim.Track,
			sim.State, startTs, endTs, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
		if err != nil {
			return err
		}
	} else {
		_, err = pstmt.Exec(sim.ID, sim.DurationInMinutes, sim.SampleRate, sim.GrandPrix, sim.Track,
			sim.State, startTs, nil, sim.PercentComplete, sim.FinalStatusCode, sim.FinalStatusMessage)
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

func (sim Simulation) UpdateEndTimestamp() error {

	var t time.Time

	sqlStatement := `UPDATE simulation SET end_timestamp = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	// Re-format the timestamp to mysql format
	t, err = pbts.Timestamp(sim.EndTimestamp)
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

/*
func NewFromProto(pbsim pb.Simulation) *Simulation {

	s := new(Simulation)
	s.ID = pbsim.Uuid
	s.DurationInMinutes = pbsim.DurationInMinutes
	s.SampleRate = pbsim.SampleRate.String()
	s.GrandPrix = pbsim.GrandPrix.String()
	s.Track = pbsim.Track.String()
	State
	StartTimestamp
	EndTimestamp
	PercentComplete
	FinalStatusCode
	FinalStatusMessage

}
*/
