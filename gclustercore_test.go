package gclustercore

import (
	"encoding/json"
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestLaunch(t *testing.T) {

	tconf := &TestConfiguration{
		GitRepo:        "https://$(GIT_USERNAME):$(GIT_PASSWORD)@github.com/sellerin/gatling-simulation.git",
		Revision:       "master",
		SimulationName: "c2gwebaws.C2gwebSimulation",
		Data:           "{\"param1\":\"value1\",\"param2\":\"value2\"}",
		NbInjectords:   2,
		NbVirtualUsers: 2,
		Duration:       300,
		Ramp:           10,
	}

	id := LaunchTest(tconf, NamespaceDev)
	fmt.Printf("Gatling test started. Id: %s\n", id)

}

func TestGetStatus(t *testing.T) {

	id, _ := uuid.FromString("e9b728fd-5f50-46d3-bd5f-db0f82a0c711")

	status := GetStatus(&id, NamespaceDev)

	status2B, _ := json.Marshal(*status)
	fmt.Println(string(status2B))

}

func TestDeleteJobs(t *testing.T) {
	DeleteJobs(NamespaceDev)
}
