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
	// checkServiceHealthCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkServiceHealthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(startSimulationCmd)

	//checkServiceHealthCmd.Flags().StringP("name", "n", "", "run health check on a FOTAAS service by name")
	//checkServiceHealthCmd.Flags().BoolP("all", "a", false, "run health check on all FOTAAS services")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}
}

var startSimulationCmd = &cobra.Command{
	Use:   "startSimulation",
	Short: "Starts a pre-defined FOTAAS simulation.",
	Long:  `Starts a FOTAAS simulation that will generate and persist telemetry data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := startSimulation()
		if err != nil {
			log.Printf("start simulation service call failed with error: %v", err)
		} else {
			log.Printf("start simulation status code   : %v", resp.ServiceStatus.Code)
			log.Printf("start simulation status message: %s", resp.ServiceStatus.Message)
		}
		return nil
	},
}

func startSimulation() (*api.RunSimulationResponse, error) {

	var simulationSvcEndpoint string
	var sb strings.Builder
	var resp *api.RunSimulationResponse
	var req api.RunSimulationRequest
	var simID string

	simMemberMap := make(map[string]*api.SimulationMember)
	simID = uuid.New().String()

	simMemberID := uuid.New().String()
	simMember1 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_HAAS,
		CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = &simMember1

	simMemberID = uuid.New().String()
	simMember2 := api.SimulationMember{Uuid: simMemberID, SimulationUuid: simID, Constructor: api.Constructor_MERCEDES,
		CarNumber: 44, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = &simMember2

	sim := api.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GranPrix: api.GranPrix_UNITED_STATES,
		Track: api.Track_AUSTIN, SimulationMemberMap: simMemberMap}

	req.Simulation = &sim

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

	resp, err = client.RunSimulation(ctx, &req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
