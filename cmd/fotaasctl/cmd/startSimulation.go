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

	pb "github.com/bburch01/FOTAAS/api"
	uid "github.com/google/uuid"
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
		log.Fatalf("godotenv error: %v", err)
	}
}

var startSimulationCmd = &cobra.Command{
	Use:   "startSimulation",
	Short: "Starts a pre-defined FOTAAS simulation.",
	Long:  `Starts a FOTAAS simulation that will generate and persist telemetry data`,
	RunE: func(cmd *cobra.Command, args []string) error {

		/*
			chkall, _ := cmd.Flags().GetBool("all")
			if chkall {
				checkAll()
			} else {
				svcname, _ := cmd.Flags().GetString("name")
				if svcname == "" {
					checkAll()
				} else {
					resp, err := checkByName(svcname)
					if err != nil {
						//log.Printf("%v service health check failed with error: %v", svcname, err)
						log.Printf("%v service health check failed with error: %v", svcname, err)
					} else {
						log.Printf("%v service health status code   : %v", svcname, resp.ServerStatus.Code)
						log.Printf("%v service health status message: %s", svcname, resp.ServerStatus.Message)
					}
				}
			}
		*/

		resp, err := startSimulation()
		if err != nil {
			log.Printf("start simulation failed with error: %v", err)
		} else {

			for _, v := range resp.ServerStatus {
				//log.Printf("uuid: %v status code: %v status msg: %v", i, v.Code, v.Message)
				log.Printf("start simulation status code   : %v", v.Code)
				log.Printf("start simulation status message: %s", v.Message)
			}

		}

		return nil
	},
}

func startSimulation() (*pb.RunSimulationResponse, error) {

	var simulationSvcEndpoint string
	var sb strings.Builder

	var resp *pb.RunSimulationResponse
	var req pb.RunSimulationRequest
	var simID string

	simmap := make(map[string]*pb.Simulation)

	simID = uid.New().String()
	sim := pb.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: pb.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: pb.SimulationRateMultiplier_X2, GrandPrix: pb.GrandPrix_UNITED_STATES,
		Track: pb.Track_AUSTIN, Constructor: pb.Constructor_HAAS, CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}
	simmap[simID] = &sim

	simID = uid.New().String()
	sim = pb.Simulation{Uuid: simID, DurationInMinutes: int32(1), SampleRate: pb.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: pb.SimulationRateMultiplier_X2, GrandPrix: pb.GrandPrix_UNITED_STATES,
		Track: pb.Track_AUSTIN, Constructor: pb.Constructor_MERCEDES, CarNumber: 44, ForceAlarm: false, NoAlarms: true,
	}
	simmap[simID] = &sim

	req.SimulationMap = simmap

	sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
	simulationSvcEndpoint = sb.String()

	conn, err := grpc.Dial(simulationSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	// TODO: determine what is the appropriate deadline for transmit requests, possibly scaling
	// based on the size of SimulationMap.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client pb.SimulationServiceClient

	client = pb.NewSimulationServiceClient(conn)

	// Need to kick-off the simulation and return a response immediately
	// Use a go routine call for client.RunSimulation
	// Status and results of the simulation will need to be persisted to
	// the simulation service db where clients can check for status/results
	resp, err = client.RunSimulation(ctx, &req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}
