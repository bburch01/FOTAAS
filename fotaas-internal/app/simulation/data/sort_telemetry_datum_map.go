package data

func sortTelemetryDatumMapByTimestamp() {

	/*
		    var td pb.TelemetryData
			var tstamp time.Time
			var err error
			var ss []kv

			for _, v := range td.TelemetryDatumMap {
				if v.Description == pb.TelemetryDatumDescription_BRAKE_TEMP_FL {
					//if v.Description == pb.TelemetryDatumDescription_TIRE_PRESSURE_FL {
					if tstamp, err = ts.Timestamp(v.Timestamp); err != nil {
						t.Error("failed to convert google.protobuf.timestamp to time.Time with error: ", err)
					}
					log.Printf("datum uuid: %v desc: %v unit: %v timestamp: %v value: %v", v.Uuid, v.Description.String(),
						v.Unit.String(), tstamp, v.Value)
				}
			}

			for k, v := range td.TelemetryDatumMap {
				if v.Description == pb.TelemetryDatumDescription_BRAKE_TEMP_FL {
					//if v.Description == pb.TelemetryDatumDescription_TIRE_PRESSURE_FL {
					ss = append(ss, kv{k, v})
				}
			}

			log.Printf("len ss: %v", len(ss))

			log.Println("\n\nSorted By Second ASCENDING:")
			sort.Slice(ss, func(i, j int) bool {
				return ss[i].Value.Timestamp.Seconds < ss[j].Value.Timestamp.Seconds
			})

			for _, kv := range ss {
				//fmt.Printf("%s, %v\n", kv.Key, kv.Value)
				fmt.Printf("\nuuid: %v desc: %v timestamp: %v value: %v", kv.Value.Uuid, kv.Value.Description, kv.Value.Timestamp, kv.Value.Value)
			}

			fmt.Print("\n\n")

			/*
				log.Println("\n\nSorted By Second DESCENDING:")

				sort.Slice(ss, func(i, j int) bool {
					return ss[i].Value.Timestamp.Seconds > ss[j].Value.Timestamp.Seconds
				})

				for _, kv := range ss {
					fmt.Printf("%s, %v\n", kv.Key, kv.Value)
				}
	*/

}
