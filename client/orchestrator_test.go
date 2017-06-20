package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/MorpheoOrg/go-packages/client"
	"github.com/MorpheoOrg/go-packages/common"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// Unit Tests: Uplet status field updates
func TestUpletUpdateNormal(t *testing.T) {
	t.Parallel()

	ts, apiClient := orchestratorAPIUpdateMock(t, http.StatusOK)
	defer ts.Close()

	err := apiClient.UpdateUpletStatus(common.TypeLearnUplet, common.TaskStatusPending, uuid.NewV4())
	assert.Nil(t, err)
}

func TestUpletUpdateBadType(t *testing.T) {
	t.Parallel()

	ts, apiClient := orchestratorAPIUpdateMock(t, http.StatusOK)
	defer ts.Close()

	err := apiClient.UpdateUpletStatus("weirduplet", common.TaskStatusDone, uuid.NewV4())
	assert.Error(t, err)
}

func TestUpletUpdateBadTaskStatus(t *testing.T) {
	t.Parallel()

	ts, apiClient := orchestratorAPIUpdateMock(t, http.StatusOK)
	defer ts.Close()

	err := apiClient.UpdateUpletStatus(common.TypePredUplet, "invalid", uuid.NewV4())
	assert.Error(t, err)
}

func TestUpletUpdateUnavailableServer(t *testing.T) {
	t.Parallel()

	apiClient := &client.OrchestratorAPI{
		Hostname: "oblivion",
		Port:     666,
	}

	err := apiClient.UpdateUpletStatus(common.TypePredUplet, common.TaskStatusDone, uuid.NewV4())
	assert.Error(t, err)
}

func TestUpletUpdateBadHTTPStatus(t *testing.T) {
	t.Parallel()

	ts, apiClient := orchestratorAPIUpdateMock(t, http.StatusTeapot)
	defer ts.Close()

	err := apiClient.UpdateUpletStatus(common.TypePredUplet, common.TaskStatusDone, uuid.NewV4())
	assert.Error(t, err)
}

type statuplet struct {
	status string
}

func orchestratorAPIUpdateMock(t *testing.T, statusCode int) (testServer *httptest.Server, apiClient client.Orchestrator) {
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload statuplet
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.Nil(t, err)
		w.WriteHeader(statusCode)
	}))

	testURL, _ := url.Parse(testServer.URL)
	host := testURL.Hostname()
	port, _ := strconv.Atoi(testURL.Port())

	apiClient = &client.OrchestratorAPI{
		Hostname: host,
		Port:     port,
	}
	return
}

// Unit Tests: Performance upload

// Uplet perf upload success
func TestPerfUploadNormal(t *testing.T) {
	t.Parallel()

	ts, apiClient := orchestratorAPIPostPerfMock(t, http.StatusOK)
	defer ts.Close()

	err := apiClient.PostLearnResult(uuid.NewV4(), client.Perfuplet{
		Status:    "wesh",
		Perf:      1.0,
		TrainPerf: map[string]float64{},
		TestPerf:  map[string]float64{},
	})
	assert.Nil(t, err)
}

// Test badly formatted JSON

// Test non existing learnuplet

func orchestratorAPIPostPerfMock(t *testing.T, statusCode int) (testServer *httptest.Server, apiClient client.Orchestrator) {
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload statuplet
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.Nil(t, err)
		w.WriteHeader(statusCode)
	}))

	testURL, _ := url.Parse(testServer.URL)
	host := testURL.Hostname()
	port, _ := strconv.Atoi(testURL.Port())

	apiClient = &client.OrchestratorAPI{
		Hostname: host,
		Port:     port,
	}
	return
}
