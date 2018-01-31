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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	// "github.com/MorpheoOrg/morpheo-go-packages/common"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
)

// Peer describes Morpheo Peer's API
type Peer interface {
	Query(queryFcn string, queryArgs []string) ([]byte, error)
	Invoke(txFcn string, txArgs []string) (string, []byte, error)

	RegisterItem(itemType, storageAddress string, problemKeys []string, itemName string) (string, []byte, error)
	RegisterProblem(storageAddress string, sizeTrainDataset int, testData []string) (string, []byte, error)
	SetUpletWorker(upletKey, worker string) (string, []byte, error)
	QueryStatusLearnuplet(status string) ([]byte, error)
	ReportLearn(upletKey, status string, perf float64, trainPerf, testPerf map[string]float64) (string, []byte, error)
}

// ============================================================================
// Peer API: Fabric Hyperledger implementation of the Peer interface
// ============================================================================

// PeerAPI describes the Fabric hyperledger peer implementation
type PeerAPI struct {
	sdk        *fabapi.FabricSDK
	ConfigFile string

	OrgID       string
	ChannelID   string
	ChaincodeID string

	ConnectEventHub bool
}

// NewPeerAPI create a new PeerAPI object
func NewPeerAPI(configFile, orgID, channelID, chaincodeID string) (*PeerAPI, error) {
	// Create SDK
	sdkOptions := fabapi.Options{
		ConfigFile: configFile,
	}
	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		return nil, fmt.Errorf("[peer-API] Error creating SDK: %s", err)
	}
	// Check that peer is available
	chClient, err := sdk.NewChannelClient(channelID, "Admin")
	if err != nil {
		return nil, fmt.Errorf("[peer-api] Failed to create new channel client: %s", err)
	}
	chClient.Close()

	return &PeerAPI{
		sdk:             sdk,
		ConfigFile:      configFile,
		OrgID:           orgID,
		ChannelID:       channelID,
		ChaincodeID:     chaincodeID,
		ConnectEventHub: true,
	}, nil
}

// ============================================================================
// Basic Functions: Query and Invoke
// ============================================================================

// Query performs a query on the Fabric Peer
func (s *PeerAPI) Query(fcn string, args []string) ([]byte, error) {
	// Create Channel Client
	chClient, err := s.sdk.NewChannelClient(s.ChannelID, "Admin")
	if err != nil {
		return nil, fmt.Errorf("[peer-api] Failed to create new channel client: %s", err)
	}
	defer chClient.Close()

	// Format args
	argsBytes := [][]byte{}
	for _, arg := range args {
		argsBytes = append(argsBytes, []byte(arg))
	}

	// Make query
	query, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: s.ChaincodeID, Fcn: fcn, Args: argsBytes})
	if err != nil {
		return nil, fmt.Errorf("[peer-api] Failed to Query (Fcn: %s, Args: %s): %s", fcn, args, err)
	}
	return query, nil
}

// Invoke performs a invoke on the Fabric Peer
func (s *PeerAPI) Invoke(fcn string, args []string) (string, []byte, error) {
	// Create Channel Client
	chClient, err := s.sdk.NewChannelClient(s.ChannelID, "Admin")
	if err != nil {
		return "", nil, fmt.Errorf("[peer-api] Failed to create new channel client: %s", err)
	}
	defer chClient.Close()

	// Format Args
	argsBytes := [][]byte{}
	for _, arg := range args {
		argsBytes = append(argsBytes, []byte(arg))
	}

	// Make query
	txID, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: s.ChaincodeID, Fcn: fcn, Args: argsBytes})
	if err != nil {
		return "", nil, fmt.Errorf("[peer-api] Failed to Execute transaction (Fcn: %s, Args: %s): %s", fcn, args, err)
	}
	return txID.ID, txID.Nonce, nil
}

// ============================================================================
// Register Functions
// ============================================================================

// RegisterItem registers an item
func (s *PeerAPI) RegisterItem(itemType, storageAddress string, problemKeys []string, itemName string) (string, []byte, error) {
	return s.Invoke("registerItem", []string{itemType, storageAddress, strings.Join(problemKeys, ","), itemName})
}

// RegisterProblem registers a problem
func (s *PeerAPI) RegisterProblem(storageAddress string, sizeTrainDataset int, testData []string) (string, []byte, error) {
	args := []string{storageAddress, strconv.Itoa(sizeTrainDataset), strings.Join(testData, ",")}
	return s.Invoke("registerProblem", args)
}

// ============================================================================
// Compute Functions
// ============================================================================

// QueryStatusLearnuplet queries the learnuplet by status
func (s *PeerAPI) QueryStatusLearnuplet(status string) ([]byte, error) {
	return s.Query("queryStatusLearnuplet", []string{status})
}

// SetUpletWorker invokes the function setUpletWorker
func (s *PeerAPI) SetUpletWorker(upletKey, worker string) (string, []byte, error) {
	return s.Invoke("setUpletWorker", []string{upletKey, worker})
}

// ReportLearn reports the output of a learning task
func (s *PeerAPI) ReportLearn(upletKey, status string, perf float64, trainPerf, testPerf map[string]float64) (string, []byte, error) {
	// Format Args
	perfArg := strconv.FormatFloat(perf, 'e', -1, 32)
	trainPerfArg, err := json.Marshal(trainPerf)
	if err != nil {
		return "", nil, fmt.Errorf("[peer-api] Failed to marshal trainPerf: %s", err)
	}
	testPerfArg, err := json.Marshal(testPerf)
	if err != nil {
		return "", nil, fmt.Errorf("[peer-api] Failed to marshal testPerf: %s", err)
	}

	// Execute Transaction
	return s.Invoke("reportLearn", []string{upletKey, status, perfArg, string(trainPerfArg), string(testPerfArg)})
}

// ============================================================================
// Peer MOCK
// ============================================================================

// PeerMock describes a mock implementation of Peer
type PeerMock struct {
}

// Query performs a query
func (s *PeerMock) Query(queryFcn string, queryArgs []string) ([]byte, error) {
	return nil, nil
}

// Invoke performs an invoke
func (s *PeerMock) Invoke(txFcn string, txArgs []string) (string, []byte, error) {
	return "", nil, nil
}

// RegisterItem registers an item
func (s *PeerMock) RegisterItem(itemType, storageAddress string, problemKeys []string, itemName string) (string, []byte, error) {
	return "", nil, nil
}

// RegisterProblem registers a problem
func (s *PeerMock) RegisterProblem(storageAddress string, sizeTrainDataset int, testData []string) (string, []byte, error) {
	return "", nil, nil
}

// SetUpletWorker invokes the function setUpletWorker
func (s *PeerMock) SetUpletWorker(upletKey, worker string) (string, []byte, error) {
	return "", nil, nil
}

// QueryStatusLearnuplet queries the learnuplet by status
func (s *PeerMock) QueryStatusLearnuplet(status string) ([]byte, error) {
	return nil, nil
}

// ReportLearn reports the output of a learning task
func (s *PeerMock) ReportLearn(upletKey, status string, perf float64, trainPerf, testPerf map[string]float64) (string, []byte, error) {
	return "", nil, nil
}
