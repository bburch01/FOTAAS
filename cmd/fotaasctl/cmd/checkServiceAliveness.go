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

	"github.com/bburch01/FOTAAS/api"
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

	rootCmd.AddCommand(checkServiceAlivenessCmd)

	checkServiceAlivenessCmd.Flags().StringP("name", "n", "", "run health check on a FOTAAS service by name")
	checkServiceAlivenessCmd.Flags().BoolP("all", "a", false, "run health check on all FOTAAS services")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}
}

var checkServiceAlivenessCmd = &cobra.Command{
	Use:   "checkServiceAliveness",
	Short: "FOTAAS services health check.",
	Long:  `Runs a health check on one or more of the FOTAAS services`,
	RunE: func(cmd *cobra.Command, args []string) error {

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
					log.Printf("%v service health check call failed with error: %v", svcname, err)
				} else {
					log.Printf("%v service health check response code   : %v", svcname, resp.Details.Code)
					log.Printf("%v service health check response message: %s", svcname, resp.Details.Message)
				}
			}
		}

		return nil
	},
}

func checkByName(svcname string) (*api.AlivenessCheckResponse, error) {

	var svcEndpoint string
	var resp *api.AlivenessCheckResponse
	var sb strings.Builder

	switch svcname {
	case "telemetry":
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("TELEMETRY_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "analysis":
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("ANALYSIS_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "simulation":
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("SIMULATION_SERVICE_PORT"))
		svcEndpoint = sb.String()
	case "status":
		sb.WriteString(os.Getenv("STATUS_SERVICE_HOST"))
		sb.WriteString(":")
		sb.WriteString(os.Getenv("STATUS_SERVICE_PORT"))
		svcEndpoint = sb.String()
	default:
		return resp, errors.New("invalid service name, valid service names are telemetry, analysis, simulation, status")
	}

	conn, err := grpc.Dial(svcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	switch svcname {
	case "telemetry":
		client := api.NewTelemetryServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "analysis":
		client := api.NewAnalysisServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "simulation":
		client := api.NewSimulationServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "status":
		client := api.NewSystemStatusServiceClient(conn)
		resp, err = client.AlivenessCheck(ctx, &api.AlivenessCheckRequest{})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid service name, valid service names are: telemetry, analysis, simulation, status")
	}

	return resp, nil
}

func checkAll() {

	resp, err := checkByName("telemetry")

	if err != nil {
		log.Printf("telemetry service health check call failed with error: %v", err)
	} else {
		log.Printf("telemetry service health check response code   : %v", resp.Details.Code)
		log.Printf("telemetry service health check response message: %s", resp.Details.Message)
	}

	resp, err = checkByName("analysis")
	if err != nil {
		log.Printf("analysis service health check call failed with error: %v", err)
	} else {
		log.Printf("analysis service health check response code   : %v", resp.Details.Code)
		log.Printf("analysis service health check response message: %s", resp.Details.Message)
	}

	resp, err = checkByName("simulation")
	if err != nil {
		log.Printf("simulation service health check call failed with error: %v", err)
	} else {
		log.Printf("simulation service health check response code   : %v", resp.Details.Code)
		log.Printf("simulation service health check response message: %s", resp.Details.Message)
	}

	resp, err = checkByName("status")
	if err != nil {
		log.Printf("status service health check call failed with error: %v", err)
	} else {
		log.Printf("status service health check response code   : %v", resp.Details.Code)
		log.Printf("status service health check response message: %s", resp.Details.Message)
	}
}
