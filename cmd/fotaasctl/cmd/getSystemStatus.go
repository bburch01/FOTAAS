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
	"github.com/briandowns/spinner"
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

	rootCmd.AddCommand(getSystemStatusCmd)

	//checkServiceAlivenessCmd.Flags().StringP("name", "n", "", "run health check on a FOTAAS service by name")
	//checkServiceAlivenessCmd.Flags().BoolP("all", "a", false, "run health check on all FOTAAS services")

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err := godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}
}

var getSystemStatusCmd = &cobra.Command{
	Use:   "getSystemStatus",
	Short: "Retreives a detailed status report from the FOTAAS system.",
	Long:  `Retreives a detailed status report from the FOTAAS system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := getSystemStatus()
		if err != nil {
			log.Printf("get system status service call failed with error: %v", err)
		} else {
			log.Printf("get system status response code   : %v", resp.Details.Code)
			log.Printf("get system status response message: %s", resp.Details.Message)
			log.Printf("telemetry service aliveness test  : %v", resp.SystemStatusReport.TelemetryServiceAliveness.String())
			log.Printf("analysis service aliveness test   : %v", resp.SystemStatusReport.AnalysisServiceAliveness.String())
			log.Printf("simulation service aliveness test : %v", resp.SystemStatusReport.SimulationServiceAliveness.String())
			log.Printf("start simulation test             : %v", resp.SystemStatusReport.StartSimulation.String())
			log.Printf("poll for simulation complete test : %v", resp.SystemStatusReport.PollForSimulationComplete.String())
			log.Printf("retrieve simulation data test     : %v", resp.SystemStatusReport.RetrieveSimulationData.String())
			log.Printf("simulation data analysis test     : %v", resp.SystemStatusReport.SimulationDataAnalysis.String())
		}
		return nil
	},
}

func getSystemStatus() (*api.GetSystemStatusResponse, error) {

	var statusSvcEndpoint string
	var sb strings.Builder
	var resp *api.GetSystemStatusResponse

	req := api.GetSystemStatusRequest{}

	sb.WriteString(os.Getenv("STATUS_SERVICE_HOST"))
	sb.WriteString(":")
	sb.WriteString(os.Getenv("STATUS_SERVICE_PORT"))
	statusSvcEndpoint = sb.String()

	conn, err := grpc.Dial(statusSvcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// TODO: determine what the appropriate deadline should be for this service call.
	clientDeadline := time.Now().Add(time.Duration(300) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	var client = api.NewSystemStatusServiceClient(conn)

	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond)
	s.Prefix = "system status test in progress, this could take a few minutes... "
	s.Start()

	resp, err = client.GetSystemStatus(ctx, &req)
	if err != nil {
		return nil, err
	}

	s.Stop()

	return resp, nil
}
