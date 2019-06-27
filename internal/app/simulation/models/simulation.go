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

func (sim *Simulation) Persist() error {

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

	return nil

}

func (sim Simulation) UpdateState() error {

	sqlStatement := `UPDATE simulation SET state = ? WHERE id = ?`

	pstmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer pstmt.Close()

	_, err = pstmt.Exec(sim.ID, sim.State)
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

	_, err = pstmt.Exec(sim.ID, endTs)
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

	_, err = pstmt.Exec(sim.ID, sim.PercentComplete)
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

	_, err = pstmt.Exec(sim.ID, sim.FinalStatusCode)
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

	_, err = pstmt.Exec(sim.ID, sim.FinalStatusMessage)
	if err != nil {
		return err
	}

	return nil

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
