package analysis

import (
	"fmt"
	"log"
	"testing"
	"time"

	ipbts "github.com/bburch01/FOTAAS/internal/pkg/protobuf/timestamp"

	"github.com/bburch01/FOTAAS/api"
	"github.com/bburch01/FOTAAS/internal/app/analysis/models"
	"github.com/joho/godotenv"
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

func TestExtractAlarmAnalysisData(t *testing.T) {

	var err error
	var startTime, endTime time.Time

	req := new(api.GetAlarmAnalysisRequest)

	startTime, err = time.Parse(time.RFC3339, "2019-07-14T00:00:00Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	endTime, err = time.Parse(time.RFC3339, "2019-07-18T23:59:59Z")
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	req.DateRangeBegin, err = ipbts.TimestampProto(startTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}
	req.DateRangeEnd, err = ipbts.TimestampProto(endTime)
	if err != nil {
		t.Error("failed to create start timestamp with error: ", err)
		t.FailNow()
	}

	req.Simulated = true

	data, err := ExtractAlarmAnalysisData(req)
	if err != nil {
		t.Error("failed to extract alarm analysis data with error: ", err)
		t.FailNow()
	}

	if data != nil {
		logger.Debug(fmt.Sprintf("simulated: %v", data.Simulated))
		logger.Debug(fmt.Sprintf("date range begin: %v", ipbts.TimestampString(data.DateRangeBegin)))
		logger.Debug(fmt.Sprintf("date range end: %v", ipbts.TimestampString(data.DateRangeEnd)))

		for _, ac := range data.AlarmCounts {
			logger.Debug(fmt.Sprintf("constructor: %v car number: %v low alarm count: %v high alarm count: %v",
				ac.Constructor.String(), ac.CarNumber, ac.LowAlarmCount, ac.HighAlarmCount))
		}
	} else {
		logger.Debug("no alarm analysis data found")
	}

}
