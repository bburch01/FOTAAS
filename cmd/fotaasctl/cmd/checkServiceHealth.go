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

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "github.com/bburch01/FOTAAS/api"
)

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkServiceHealthCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkServiceHealthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(checkServiceHealthCmd)

	checkServiceHealthCmd.Flags().StringP("name", "n", "", "run health check on a FOTAAS service by name")
	checkServiceHealthCmd.Flags().BoolP("all", "a", false, "run health check on all FOTAAS services")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv error: %v", err)
	}
}

// checkServiceHealthCmd represents the checkServiceHealth command
var checkServiceHealthCmd = &cobra.Command{
	Use:   "checkServiceHealth",
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
					//log.Printf("%v service health check failed with error: %v", svcname, err)
					log.Printf("%v service health check failed with error: %v", svcname, err)
				} else {
					log.Printf("%v service health status code   : %v", svcname, resp.ServerStatus.Code)
					log.Printf("%v service health status message: %s", svcname, resp.ServerStatus.Message)
				}
			}
		}

		return nil
	},
}

func checkByName(svcname string) (*pb.HealthCheckResponse, error) {

	var svcaddr string
	var resp *pb.HealthCheckResponse

	switch svcname {
	case "telemetry":
		serviceaddress := []string{os.Getenv("TELEMETRY_SERVICE_HOST"), ":", os.Getenv("TELEMETRY_SERVICE_PORT")}
		svcaddr = strings.Join(serviceaddress, "")

	case "analysis":
		serviceaddress := []string{os.Getenv("ANALYSIS_SERVICE_HOST"), ":", os.Getenv("ANALYSIS_SERVICE_PORT")}
		svcaddr = strings.Join(serviceaddress, "")

	case "simulation":
		serviceaddress := []string{os.Getenv("SIMULATION_SERVICE_HOST"), ":", os.Getenv("SIMULATION_SERVICE_PORT")}
		svcaddr = strings.Join(serviceaddress, "")

	default:
		return resp, errors.New("invalid service name, valid service names are telemetry, analysis, and simulation")
	}

	conn, err := grpc.Dial(svcaddr, grpc.WithInsecure())
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)

	clientDeadline := time.Now().Add(time.Duration(1) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	switch svcname {
	case "telemetry":
		client := pb.NewTelemetryServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &pb.HealthCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "analysis":
		client := pb.NewAnalysisServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &pb.HealthCheckRequest{})
		if err != nil {
			return nil, err
		}
	case "simulation":
		client := pb.NewSimulationServiceClient(conn)
		resp, err = client.HealthCheck(ctx, &pb.HealthCheckRequest{})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid service name, valid service names are: telemetry, analysis, simulation")
	}

	return resp, nil
}

func checkAll() {

	resp, err := checkByName("telemetry")
	if err != nil {
		log.Printf("telemetry service health check failed with error: %v", err)
	} else {
		log.Printf("telemetry service health status code   : %v", resp.ServerStatus.Code)
		log.Printf("telemetry service health status message: %s", resp.ServerStatus.Message)
	}

	resp, err = checkByName("analysis")
	if err != nil {
		log.Printf("analysis service health check failed with error: %v", err)
	} else {
		log.Printf("analysis service health status code   : %v", resp.ServerStatus.Code)
		log.Printf("analysis service health status message: %s", resp.ServerStatus.Message)
	}

	resp, err = checkByName("simulation")
	if err != nil {
		log.Printf("simulation health check failed with error: %v", err)
	} else {
		log.Printf("simulation service health status code   : %v", resp.ServerStatus.Code)
		log.Printf("simulation service health status message: %s", resp.ServerStatus.Message)
	}
}
