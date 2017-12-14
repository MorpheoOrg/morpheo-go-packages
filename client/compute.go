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
	"net/http"

	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

// Compute HTTP API routes
const (
	ComputeLearnupletRoute = "learn"
	ComputePredupletRoute  = "pred"
)

// Compute describes Morpheo's compute API
type Compute interface {
	PostLearnuplet(learnuplet common.Learnuplet) error
	PostPreduplet(preduplet common.Preduplet) error
}

// ComputeAPI is a wrapper around our compute API
type ComputeAPI struct {
	Compute

	Hostname string
	Port     int
	// User     string
	// Password string
}

// PostLearnuplet forwards a JSON-formatted learn result to the compute HTTP API
func (s *ComputeAPI) PostLearnuplet(learnuplet common.Learnuplet) error {
	return s.postJSONData(ComputeLearnupletRoute, learnuplet)
}

// PostPreduplet forwards a JSON-formatted pred result to the compute HTTP API
func (s *ComputeAPI) PostPreduplet(preduplet common.Preduplet) error {
	return s.postJSONData(ComputePredupletRoute, preduplet)
}

func (s *ComputeAPI) postJSONData(route string, resource interface{}) error {
	url := fmt.Sprintf("http://%s:%d/%s", s.Hostname, s.Port, route)

	dataBytes, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("[compute-api] Error building POST request against %s: Error marshaling to JSON: %+v", url, resource)
	}
	data := bytes.NewReader(dataBytes)

	req, err := http.NewRequest(http.MethodPost, url, data)
	if err != nil {
		return fmt.Errorf("[compute-api] Error building result POST request against %s: %s", url, err)
	}
	// req.SetBasicAuth(s.User, s.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[compute-api] Error performing result POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("[compute-api] Unexpected status code (%s): result POST request against %s, \nBody: %s", resp.Status, url, string(body))
	}
	return nil
}
