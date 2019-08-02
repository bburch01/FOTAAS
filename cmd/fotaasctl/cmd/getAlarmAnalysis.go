// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkServiceAlivenessCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkServiceAlivenessCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(getAlarmAnalysisCmd)
	getAlarmAnalysisCmd.Flags().StringP("start-date", "s", "", "alarm analysis start date (yyyy-mm-dd)")
	getAlarmAnalysisCmd.Flags().StringP("end-date", "e", "", "alarm analysis end date (yyyy-mm-dd)")
	getAlarmAnalysisCmd.Flags().StringP("constructor", "c", "", "constructor (e.g. MERCEDES")
	getAlarmAnalysisCmd.Flags().Int32P("car-number", "n", -1, "car number (e.g. 44)")
	getAlarmAnalysisCmd.Flags().BoolP("simulated", "i", false, "get alarm analysis for simulated data")
	getAlarmAnalysisCmd.Flags().StringP("simulation-id", "d", "", "get alarm analysis for a specific simulation uuid")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}
}

var getAlarmAnalysisCmd = &cobra.Command{
	Use:   "getAlarmAnalysis",
	Short: "Produces a telemetry alarm analysis for a given date range.",
	Long: `Produces a telemetry alarm analysis for a given date range. A constructor and car number
	 can be specified to produce a report specific to that constructor and car.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		startDate, _ := cmd.Flags().GetString("start-date")
		if startDate == "" {
			return errors.New("start-date must be specified, format is yyyy-mm-dd")
		}
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			return errors.New("invalid start-date specified, format is yyyy-mm-dd")
		}

		endDate, _ := cmd.Flags().GetString("end-date")
		if endDate == "" {
			return errors.New("end-date must be specified, format is yyyy-mm-dd")
		}
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			return errors.New("invalid end-date specified, format is yyyy-mm-dd")
		}

		simulated, _ := cmd.Flags().GetBool("simulated")

		simID, _ := cmd.Flags().GetString("simulation-id")

		if simID != "" {
			if _, err := uuid.Parse(simID); err != nil {
				log.Printf("invalid simulation id: %v", err)
				return nil
			}
		}

		constructor, _ := cmd.Flags().GetString("constructor")
		if constructor != "" {

			constructor = strings.ToUpper(constructor)

			constructorOrdinal, ok := api.Constructor_value[constructor]
			if !ok {
				return errors.New("invalid constructor specified, valid constructors are: alpha_romeo, ferrari, haas, mclaren, mercedes, racing_point, red_bull_racing, scuderia_toro_roso, williams")
			}

			carNumber, err := cmd.Flags().GetInt32("car-number")

			if err != nil || carNumber < 0 {
				return errors.New("car-number must be specified and must be greater than or equal to 0")
			}

			resp, err := getConstructorAlarmAnalysis(startDate, endDate, simulated, simID, constructorOrdinal, carNumber)
			if err != nil {
				return err
			}

			log.Printf("analysis service response code: %v", resp.Details.Code.String())
			log.Printf("analysis service response message: %v", resp.Details.Message)

			data := resp.ConstructorAlarmAnalysisData
			if data != nil {
				log.Printf("simulated: %v", data.Simulated)

				log.Printf("date range begin: %v", ipbts.TimestampString(data.DateRangeBegin))
				log.Printf("date range end: %v", ipbts.TimestampString(data.DateRangeEnd))
				log.Printf("constructor: %v", data.Constructor)
				log.Printf("car number: %v", data.CarNumber)

				for _, ac := range data.AlarmCounts {
					log.Printf("description: %v low alarm count: %v high alarm count: %v",
						ac.DatumDescription.String(), ac.LowAlarmCount, ac.HighAlarmCount)
				}
			}

		} else {

			resp, err := getAlarmAnalysis(startDate, endDate, simulated, simID)
			if err != nil {
				return err
			}

			log.Printf("analysis service response code: %v", resp.Details.Code.String())
			log.Printf("analysis service response message: %v", resp.Details.Message)

			data := resp.AlarmAnalysisData
			if data != nil {
				log.Printf("simulated: %v", data.Simulated)
				log.Printf("date range begin: %v", ipbts.TimestampString(data.DateRangeBegin))
				log.Printf("date range end: %v", ipbts.TimestampString(data.DateRangeEnd))

				for _, ac := range data.AlarmCounts {
					log.Printf("constructor: %v car number: %v low alarm count: %v high alarm count: %v",
						ac.Constructor.String(), ac.CarNumber, ac.LowAlarmCount, ac.HighAlarmCount)
				}
			}

		}

		return nil
	},
}

func getAlarmAnalysis(startDate string, endDate string, simulated bool, simID string) (*api.GetAlarmAnalysisResponse, error) {

	var startTime, endTime time.Time
	var err error

	req := new(api.GetAlarmAnalysisRequest)

	startTime, err = time.Parse(time.RFC3339, startDate+"T00:00:00Z")
	if err != nil {
		return nil, err
	}

	endTime, err = time.Parse(time.RFC3339, endDate+"T23:59:59Z")
	if err != nil {
		return nil, err
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		return nil, err
	}

	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		return nil, err
	}

	req.Simulated = simulated
	req.SimulationUuid = simID

	var sb strings.Builder
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
	analysisSvcEndpoint := sb.String()

	conn, err := grpc.Dial(analysisSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewAnalysisServiceClient(conn)

	var resp *api.GetAlarmAnalysisResponse
	resp, err = client.GetAlarmAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func getConstructorAlarmAnalysis(startDate string, endDate string, simulated bool, simID string,
	constructorOrdinal int32, carNumber int32) (*api.GetConstructorAlarmAnalysisResponse, error) {

	var startTime, endTime time.Time
	var err error

	req := new(api.GetConstructorAlarmAnalysisRequest)

	startTime, err = time.Parse(time.RFC3339, startDate+"T00:00:00Z")
	if err != nil {
		return nil, err
	}

	endTime, err = time.Parse(time.RFC3339, endDate+"T23:59:59Z")
	if err != nil {
		return nil, err
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		return nil, err
	}

	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		return nil, err
	}

	req.Simulated = simulated
	req.SimulationUuid = simID
	req.Constructor = api.Constructor(constructorOrdinal)
	req.CarNumber = carNumber

	var sb strings.Builder
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
	analysisSvcEndpoint := sb.String()

	conn, err := grpc.Dial(analysisSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewAnalysisServiceClient(conn)

	var resp *api.GetConstructorAlarmAnalysisResponse
	resp, err = client.GetConstructorAlarmAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
