/*
 * Copyright Morpheo Org. 2017
 *
 * contact@morpheo.co
 *
 * This software is part of the Morpheo project, an open-source machine
 * learning platform.
 *
 * This software is governed by the CeCILL license, compatible with the
 * GNU GPL, under French law and abiding by the rules of distribution of
 * free software. You can  use, modify and/ or redistribute the software
 * under the terms of the CeCILL license as circulated by CEA, CNRS and
 * INRIA at the following URL "http://www.cecill.info".
 *
 * As a counterpart to the access to the source code and  rights to copy,
 * modify and redistribute granted by the license, users are provided only
 * with a limited warranty  and the software's author,  the holder of the
 * economic rights,  and the successive licensors  have only  limited
 * liability.
 *
 * In this respect, the user's attention is drawn to the risks associated
 * with loading,  using,  modifying and/or developing or reproducing the
 * software by the user in light of its specific status of free software,
 * that may mean  that it is complicated to manipulate,  and  that  also
 * therefore means  that it is reserved for developers  and  experienced
 * professionals having in-depth computer knowledge. Users are therefore
 * encouraged to load and test the software's suitability as regards their
 * requirements in conditions enabling the security of their systems and/or
 * data to be ensured and,  more generally, to use and operate it in the
 * same conditions as regards security.
 *
 * The fact that you are presently reading this means that you have had
 * knowledge of the CeCILL license and that you accept its terms.
 */

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/MorpheoOrg/morpheo-go-packages/common"
	uuid "github.com/satori/go.uuid"
)

// Orchestrator HTTP API routes
const (
	OrchestratorStatusUpdateRoute = "worker"
	OrchestratorLearnResultRoute  = "learndone"
	OrchestratorPredResultRoute   = "preddone"
	OrchestratorAlgoRoute         = "algo"
	OrchestratorDataRoute         = "data"
	OrchestratorPredictionRoute   = "prediction"
	OrchestratorProblemRoute      = "problem"
)

var (
	// OrchestratorResultRoutes sets result routes by uplet-type
	OrchestratorResultRoutes = map[string]string{
		common.TypeLearnUplet: OrchestratorLearnResultRoute,
		common.TypePredUplet:  OrchestratorPredResultRoute,
	}
)

// Perfuplet describes the response of compute to the orchestrator
type Perfuplet struct {
	Status    string             `json:"status"`
	Perf      float64            `json:"perf"`
	TrainPerf map[string]float64 `json:"train_perf"`
	TestPerf  map[string]float64 `json:"test_perf"`
}

// Preddone describes compute requests to the orchestrator
type Preddone struct {
	Status              string    `json:"status"`
	PredictionStorageID uuid.UUID `json:"prediction_storage_uuid"`
}

// Orchestrator describes Morpheo's orchestrator API
type Orchestrator interface {
	UpdateUpletStatus(upletType, status string, upletID uuid.UUID, workerID uuid.UUID) error
	PostLearnResult(learnupletID uuid.UUID, perfuplet Perfuplet) error
	PostPredResult(predupletID uuid.UUID, preddone Preddone) error
	PostAlgo(algo common.OrchestratorAlgo) error
	PostData(data common.OrchestratorData) error
	PostPrediction(prediction common.OrchestratorPrediction) error
	PostProblem(problem common.OrchestratorProblem) error
}

// OrchestratorAPI is a wrapper around our orchestrator API
type OrchestratorAPI struct {
	Orchestrator

	Hostname string
	Port     int
	User     string
	Password string
}

// PostLearnResult forwards a JSON-formatted learn result to the orchestrator HTTP API
func (o *OrchestratorAPI) PostLearnResult(learnupletID uuid.UUID, perfuplet Perfuplet) error {
	return o.postJSONData(path.Join(OrchestratorLearnResultRoute, learnupletID.String()), perfuplet)
}

// PostPredResult forwards a JSON-formatted pred result to the orchestrator HTTP API
func (o *OrchestratorAPI) PostPredResult(predupletID uuid.UUID, preddone Preddone) error {
	return o.postJSONData(path.Join(OrchestratorPredResultRoute, predupletID.String()), preddone)
}

// PostAlgo posts a new algo to the Orchestrator
func (o *OrchestratorAPI) PostAlgo(algo common.OrchestratorAlgo) error {
	return o.postJSONData(OrchestratorAlgoRoute, algo)
}

// PostData posts a new data to the Orchestrator
func (o *OrchestratorAPI) PostData(data common.OrchestratorData) error {
	return o.postJSONData(OrchestratorDataRoute, data)
}

// PostPrediction posts a new prediction to the Orchestrator
func (o *OrchestratorAPI) PostPrediction(prediction common.OrchestratorPrediction) error {
	return o.postJSONData(OrchestratorPredictionRoute, prediction)
}

// PostProblem posts a new problem to the Orchestrator
func (o *OrchestratorAPI) PostProblem(problem common.OrchestratorProblem) error {
	return o.postJSONData(OrchestratorProblemRoute, problem)
}

// GetList retrieve the list of a resource
func (o *OrchestratorAPI) GetList(resourceType string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/%s", o.Hostname, o.Port, resourceType)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("[orchestrator-api] Error building result POST request against %s: %s", url, err)
	}
	req.SetBasicAuth(o.User, o.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[orchestrator-api] Error performing status update POST request against %s: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[orchestrator-api] Unexpected status code (%s): status update POST request against %s", resp.Status, url)
	}

	// return the body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[orchestrator-api] Error reading Body: %s", err)
	}
	return body, nil
}

// UpdateUpletStatus changes the status field of a learnuplet/preduplet
func (o *OrchestratorAPI) UpdateUpletStatus(upletType string, status string, upletID uuid.UUID, workerID uuid.UUID) error {
	// Check that arguments are valid
	if _, ok := common.ValidUplets[upletType]; !ok {
		return fmt.Errorf("[orchestrator-api] Uplet type \"%s\" is invalid. Allowed values are %s", upletType, common.ValidUplets)
	}
	if _, ok := common.ValidStatuses[status]; !ok {
		return fmt.Errorf("[orchestrator-api] Status \"%s\" is invalid. Allowed values are %s", status, common.ValidStatuses)
	}

	// TODO (orchestrator): make the orchestrator API RESTFul and get rid of this dirty logic
	var url string
	var payload []byte
	if status == common.TaskStatusPending {
		url = fmt.Sprintf("http://%s:%d/%s/%s/%s", o.Hostname, o.Port, OrchestratorStatusUpdateRoute, upletType, upletID)
		payload, _ = json.Marshal(map[string]string{"worker": workerID.String()})
	} else if status == common.TaskStatusFailed {
		url = fmt.Sprintf("http://%s:%d/%s/%s", o.Hostname, o.Port, OrchestratorResultRoutes[upletType], upletID)
		payload, _ = json.Marshal(map[string]string{"status": status})
	} else {
		return fmt.Errorf("[orchestrator-api] Status Update Error on %s %s: for now, only %s and %s statuses are supported", upletType, upletID, common.TaskStatusPending, common.TaskStatusFailed)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error building result POST request against %s: %s", url, err)
	}
	req.SetBasicAuth(o.User, o.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error performing status update POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[orchestrator-api] Unexpected status code (%s): status update POST request against %s", resp.Status, url)
	}
	return nil
}

func (o *OrchestratorAPI) postJSONData(route string, resource interface{}) error {
	url := fmt.Sprintf("http://%s:%d/%s", o.Hostname, o.Port, route)

	dataBytes, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error building POST request against %s: Error marshaling to JSON: %+v", url, resource)
	}
	data := bytes.NewReader(dataBytes)

	req, err := http.NewRequest(http.MethodPost, url, data)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error building result POST request against %s: %s", url, err)
	}
	req.SetBasicAuth(o.User, o.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error performing result POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("[orchestrator-api] Unexpected status code (%s): result POST request against %s, \nBody: %s", resp.Status, url, string(body))
	}
	return nil
}

// OrchestratorAPIMock mocks the Orchestrator API, always returning ok to update queries except for
// given "unexisting" pred/learn uplet with a given UUID
type OrchestratorAPIMock struct {
	Orchestrator

	UnexistingUplet string
}

// NewOrchestratorAPIMock returns with a mock of the Orchestrator API
func NewOrchestratorAPIMock() (s *OrchestratorAPIMock) {
	return &OrchestratorAPIMock{
		UnexistingUplet: "ea408171-0205-475e-8962-a02855767260",
	}
}

// UpdateUpletStatus returns nil except if OrchestratorAPIMock.UnexistingUpletID is passed
func (o *OrchestratorAPIMock) UpdateUpletStatus(upletType, status string, upletID uuid.UUID, workerID uuid.UUID) error {
	if upletID.String() != o.UnexistingUplet {
		log.Printf("[orchestrator-mock] Received update status from worker %s for %s %s. Status: %s", workerID, upletType, upletID, status)
		return nil
	}
	return fmt.Errorf("[orchestrator-mock][status-update] Unexisting uplet %s", upletID)
}

// PostLearnResult returns nil except if OrchestratorAPIMock.UnexistingUpletID is passed
func (o *OrchestratorAPIMock) PostLearnResult(learnupletID uuid.UUID, perfuplet Perfuplet) error {
	if learnupletID.String() != o.UnexistingUplet {
		log.Printf("[orchestrator-mock] Received learn result for learn-uplet %s: \n %+v", learnupletID, perfuplet)
		return nil
	}
	return fmt.Errorf("[orchestrator-mock][status-update] Unexisting uplet %s", learnupletID)
}

// PostPredResult returns nil except if OrchestratorAPIMock.UnexistingUpletID is passed
func (o *OrchestratorAPIMock) PostPredResult(predupletID uuid.UUID, preddone Preddone) error {
	if predupletID.String() != o.UnexistingUplet {
		log.Printf("[orchestrator-mock] Received pred result for pred-uplet %s: \n %+v", predupletID, preddone)
		return nil
	}
	return fmt.Errorf("[orchestrator-mock][status-update] Unexisting uplet %s", predupletID)
}

// PostAlgo returns nil
func (o *OrchestratorAPIMock) PostAlgo(algo common.OrchestratorAlgo) error {
	return nil
}

// PostData returns nil
func (o *OrchestratorAPIMock) PostData(data common.OrchestratorData) error {
	return nil
}

// PostPrediction returns nil
func (o *OrchestratorAPIMock) PostPrediction(prediction common.OrchestratorPrediction) error {
	return nil
}

// PostProblem returns nil
func (o *OrchestratorAPIMock) PostProblem(problem common.OrchestratorProblem) error {
	return nil
}
