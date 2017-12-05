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
	"fmt"
	"log"
	"math/rand"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
)

var (
	queryFcn  = "queryItems"
	queryArgs = [][]byte{[]byte("algo")}
	txFcn     = "registerItem"
	txArgs    = [][]byte{[]byte("algo"),
		[]byte("https://storage.morpheo.io/algo/0pa81bfc-b5f4-5ba2-b81a-b464248f02d2"),
		[]byte(fmt.Sprintf("N U M B E R : %s", rand.Int())),
	}
)

// Peer describes Morpheo's Peer API interface
// type Peer interface {
// 	UpdateUpletStatus(upletType, status string, upletID uuid.UUID, workerID uuid.UUID) error
// 	PostLearnResult(learnupletID uuid.UUID, perfuplet Perfuplet) error
// 	PostPredResult(predupletID uuid.UUID, preddone Preddone) error
// 	PostAlgo(algo common.OrchestratorAlgo) error
// 	PostData(data common.OrchestratorData) error
// 	PostPrediction(prediction common.OrchestratorPrediction) error
// 	PostProblem(problem common.OrchestratorProblem) error
// 	GetLearnuplet() error
// }

// PeerAPI
type PeerAPI struct {
	Sdk        *fabapi.FabricSDK
	ConfigFile string

	OrgID       string
	ChannelID   string
	ChaincodeID string

	ConnectEventHub bool
	// ChannelConfig   string
	// AdminUser ca.User
}

func NewPeerAPI(configFile, orgID, channelID, chaincodeID string) (*PeerAPI, error) {
	sdkOptions := fabapi.Options{
		ConfigFile: configFile,
	}
	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		return nil, fmt.Errorf("[peer-API] Error creating SDK: %s", err)
	}
	return &PeerAPI{
		Sdk:             sdk,
		ConfigFile:      configFile,
		OrgID:           orgID,
		ChannelID:       channelID,
		ChaincodeID:     chaincodeID,
		ConnectEventHub: true,
	}, nil
}

func (s *PeerAPI) GetAglo() error {
	// Create Channel Client
	chClient, err := s.Sdk.NewChannelClient(s.ChannelID, "Admin")
	if err != nil {
		return fmt.Errorf("[peer-api] Failed to create new channel client: %s", err)
	}
	defer chClient.Close()

	// Query
	query, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: s.ChaincodeID, Fcn: queryFcn, Args: queryArgs})
	if err != nil {
		return fmt.Errorf("[peer-api] Failed to query funds: %s", err)
	}

	log.Printf("\n\n*** 1st Query ***\n%s\n\n", query)

	// Execute transaction
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: s.ChaincodeID, Fcn: txFcn, Args: txArgs})
	if err != nil {
		return fmt.Errorf("[peer-api] Failed to registerItem: %s", err)
	}

	// Query new value
	newQuery, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: s.ChaincodeID, Fcn: queryFcn, Args: queryArgs})
	if err != nil {
		return fmt.Errorf("[peer-api] Failed to query funds after transaction: %s", err)
	}

	log.Printf("\n\n***  2nd QueryValue after registering ***\n%s\n\n", newQuery)

	return nil
}
