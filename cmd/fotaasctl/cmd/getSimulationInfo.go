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
	"log"
	"os"
	"strings"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	//"github.com/bburch01/FOTAAS/internal/app/simulation/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkServiceHealthCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkServiceHealthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(getSimulationInfoCmd)

	getSimulationInfoCmd.Flags().StringP("id", "i", "", "simulation id")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}
}

var getSimulationInfoCmd = &cobra.Command{
	Use:   "getSimulationInfo",
	Short: "Gets information about a FOTAAS simulation.",
	Long:  `Gets information (e.g. status, percent complete, etc) about a FOTAAS simulation (running or not).`,
	RunE: func(cmd *cobra.Command, args []string) error {

		id, _ := cmd.Flags().GetString("id")

		log.Printf("id flag value: %v", id)

		if _, err := uuid.Parse(id); err != nil {
			log.Printf("invalid simulation id: %v", err)
			return nil
		}

		resp, err := getSimulationInfo(id)
		if err != nil {
			log.Printf("get simulation info service call failed with error: %v", err)
		} else {
			log.Printf("get simulation info response code   : %v", resp.Details.Code)
			log.Printf("get simulation info response message: %s", resp.Details.Message)

			if resp.SimulationInfo != nil {

				log.Printf("\nsimulation id       : %v ", resp.SimulationInfo.Uuid)
				log.Printf("\nduration in minutes : %v ", resp.SimulationInfo.DurationInMinutes)
				log.Printf("\nsample rate         : %v ", resp.SimulationInfo.SampleRate)
				log.Printf("\ngran prix           : %v ", resp.SimulationInfo.GranPrix)
				log.Printf("\ntrack               : %v ", resp.SimulationInfo.Track)
				log.Printf("\nstate               : %v ", resp.SimulationInfo.State)
				log.Printf("\nstart timestamp     : %v ", ipbts.TimestampString(resp.SimulationInfo.StartTimestamp))
				log.Printf("\nend timestamp       : %v ", ipbts.TimestampString(resp.SimulationInfo.EndTimestamp))
				log.Printf("\npercent complete    : %v ", resp.SimulationInfo.PercentComplete)
				log.Printf("\nfinal info code   : %v ", resp.SimulationInfo.FinalStatusCode)
				log.Printf("\nfinal info message: %v ", resp.SimulationInfo.FinalStatusMessage)
				log.Print("\n")

			}

		}
		return nil
	},
}

func getSimulationInfo(simID string) (*api.GetSimulationInfoResponse, error) {

	var simulationSvcEndpoint string
	var sb strings.Builder
	var req api.GetSimulationInfoRequest
	var resp *api.GetSimulationInfoResponse

	req.SimulationUuid = simID

	sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcEndpoint = sb.String()

	conn, err := grpc.Dial(simulationSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the size of SimulationMap.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewSimulationServiceClient(conn)

	resp, err = client.GetSimulationInfo(ctx, &req)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
