package simulation

import (
	//"fmt"
	"fmt"
	"log"
	"testing"

	//"time"

	//ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	//"github.com/bburch01/FOTAAS/internal/app/simulation"
	"github.com/bburch01/FOTAAS/internal/app/simulation/models"

	//"github.com/bburch01/FOTAAS/internal/app/simulation/data"
	//"github.com/bburch01/FOTAAS/internal/app/telemetry"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	// 	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	//tel "github.com/bburch01/FOTAAS/internal/app/telemetry"
	//ts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"
)

func init() {

	var err error

	// Loads values from .env into the system.
	// NOTE: the .env file must be present in execution directory which is a
	// deployment issue that will be handled via docker/k8s in production but
	// the .env file may need to be manually copied into the execution directory
	// during testing.
	if err = godotenv.Load(); err != nil {
		log.Panicf("failed to load environment variables with error: %v", err)
	}

	if err = models.InitDB(); err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database driver with error: %v", err))
	}

}

func TestStartSimulation(t *testing.T) {

	//var sampleRateInMillis int32
	var simID string
	var simMember models.SimulationMember
	var sim models.Simulation

	simMemberMap := make(map[string]models.SimulationMember)
	simID = uuid.New().String()

	simMemberID := uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_HAAS,
		CarNumber: 8, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = simMember

	simMemberID = uuid.New().String()
	simMember = models.SimulationMember{ID: simMemberID, SimulationID: simID, Constructor: api.Constructor_MERCEDES,
		CarNumber: 44, ForceAlarm: false, NoAlarms: true,
	}
	simMemberMap[simMemberID] = simMember

	sim = models.Simulation{ID: simID, DurationInMinutes: int32(1), SampleRate: api.SampleRate_SR_1000_MS,
		SimulationRateMultiplier: api.SimulationRateMultiplier_X1, GrandPrix: api.GrandPrix_UNITED_STATES,
		Track: api.Track_AUSTIN, SimulationMembers: simMemberMap}

	// By design, StartSimulation is started asychronously. Clients will start simulation and then use
	// simulation service grpc calls to get status on running & completed simulations. This test will
	// eventually need to add a polling loop that will query the simulation service db to check on the
	// status of the test simulation. sqlmock will eventually need to be incorported so the this unit
	// test can run in the CI/CD pipeline.
	StartSimulation(&sim)

}
