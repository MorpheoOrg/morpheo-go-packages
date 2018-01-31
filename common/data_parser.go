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

package common

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ===========================================================================
// Structures
// ===========================================================================

// DataParser is used to parse all the metadata from a yaml file into structs
// and to retrieve associated data from local disk
type DataParser struct {
	Chaincode Chaincode `yaml:"chaincode"`
	Storage   Storage   `yaml:"storage"`

	// pathDataFolder is the local path where data is stored
	PathDataFolder string `yaml:"pathDataFolder"`
}

// Chaincode describes metadata used to register items to the Chaincode
// and to perform a prediction request
type Chaincode struct {
	Algo       []AlgoRegister      `yaml:"algo"`
	Data       []DataRegister      `yaml:"data"`
	Problem    []ProblemRegister   `yaml:"problem"`
	Prediction []PredictionRequest `yaml:"prediction"`
}

// Storage describes data to post resources to storage
type Storage struct {
	Algo    []Algo    `yaml:"algo"`
	Data    []Data    `yaml:"data"`
	Model   []Model   `yaml:"model"`
	Problem []Problem `yaml:"problem"`
}

// Chaincode Specific Structures
// ===========================================================================

// AlgoRegister describes the fields needed to register an algo
type AlgoRegister struct {
	StorageAddress string   `json:"storageAddress" yaml:"storageAddress"`
	ProblemKeys    []string `json:"problemKeys" yaml:"problemKeys"`
	Name           string   `json:"name" yaml:"name"`
}

// DataRegister describes the fields needed to register a data
type DataRegister struct {
	StorageAddress string   `json:"storageAddress" yaml:"storageAddress"`
	ProblemKeys    []string `json:"problemKeys" yaml:"problemKeys"`
	Name           string   `json:"name" yaml:"name"`
}

// ProblemRegister describes the fields needed to register a problem
type ProblemRegister struct {
	StorageAddress   string   `json:"storageAddress" yaml:"storageAddress"`
	TestData         []string `json:"testData" yaml:"testData"`
	SizeTrainDataset int      `json:"sizeTrainDataset" yaml:"sizeTrainDataset"`
}

// PredictionRequest describes the fields needed to request a prediction
type PredictionRequest struct {
	Data    string `json:"data" yaml:"data"`
	Problem string `json:"problem" yaml:"problem"`
}

// ===========================================================================
// Functions
// ===========================================================================

// ParseDataFromFile returns a DataParser struct holding the data in the file
// provided
func ParseDataFromFile(pathYAML string) (parser *DataParser, err error) {
	data, err := ioutil.ReadFile(pathYAML)
	if err != nil {
		return nil, fmt.Errorf("Error reading the yaml file %s: %s", pathYAML, err)
	}
	err = yaml.Unmarshal(data, &parser)
	if err != nil {
		return nil, fmt.Errorf("Error Unmarshaling the yaml file %s: %s", pathYAML, err)
	}
	return parser, nil
}

// GetData returns an io.ReaderCloser of the data specified.
// The file will be searched under pathDataFolder/dataType/*
func (s *DataParser) GetData(dataType, key string) (io.ReadCloser, error) {
	// fmt.Printf("[DEBUG] pathDataFolder: %s\n", s.PathDataFolder)
	path, err := searchFileInFolder(key, filepath.Join(s.PathDataFolder, dataType))
	if err != nil {
		return nil, fmt.Errorf("[DataParser] Error searching file %s: %s", key, err)
	}
	return os.Open(path)
}

// PrintSample prints a data sample to the console (for test purposes)
func (s *DataParser) PrintSample() {
	fmt.Println("\n\n----- Storage -----")
	fmt.Printf("\nAlgo: %+v\n", s.Storage.Algo[0])
	fmt.Printf("\nData: %+v\n", s.Storage.Data[0])
	fmt.Printf("\nModel: %+v\n", s.Storage.Model[0])
	fmt.Printf("\nProblem: %+v\n", s.Storage.Problem[0])
}

// searchFileInFolder searches for the specified file in the provided folder
// and its subdirectories, and returns the file path if successful.
func searchFileInFolder(filename, folder string) (pathFile string, err error) {
	// Check if folder exists (important, otherwise => panic Error)
	_, err = os.Stat(folder)
	if err != nil {
		return "", err
	}

	// Look for file
	err = filepath.Walk(folder, func(path string, f os.FileInfo, walkerr error) error {
		if !f.IsDir() && filename == f.Name() {
			pathFile = path
			return io.EOF
		}
		return nil
	})
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("Error walking %s: %s", folder, err)
	}
	if pathFile == "" {
		return "", fmt.Errorf("File %s not found in %s", filename, folder)
	}
	return pathFile, err
}
