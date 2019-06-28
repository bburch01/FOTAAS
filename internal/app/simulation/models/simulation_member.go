package models

type SimulationMember struct {
	ID                    string
	SimulationID          string
	Constructor           string
	CarNumber             int32
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

func (simMember SimulationMember) FindAllBySimulationID() (*[]SimulationMember, error) {

	var simMembers []SimulationMember
	var member SimulationMember

	/*
		sqlStatement := `
			SELECT * FROM simulation_member WHERE simulation_id = ?`

		pstmt, err := db.Prepare(sqlStatement)
		if err != nil {
			return nil, err
		}
		defer pstmt.Close()
	*/

	rows, err := db.Query("select id, simulation_id, constructor, car_number, force_alarm, no_alarms from simulation_member where simulation_id = ?", simMember.SimulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&member.ID, &member.SimulationID, &member.Constructor,
			&member.CarNumber, &member.ForceAlarm, &member.NoAlarms)
		if err != nil {
			return nil, err
		}
		simMembers = append(simMembers, member)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &simMembers, nil

}
