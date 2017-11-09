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

// DataParser parses all the metadata from a yaml file into
// structs and can retrieve associated data from disk
type DataParser struct {
	Orchestrator Orchestrator `yaml:"orchestrator"`
	Storage      Storage      `yaml:"storage"`

	PathDataFolder string
}

// Orchestrator holds metadata POSTable to the orchestrator
type Orchestrator struct {
	Algo       []OrchestratorAlgo       `yaml:"algo"`
	Data       []OrchestratorData       `yaml:"data"`
	Prediction []OrchestratorPrediction `yaml:"prediction"`
	Problem    []OrchestratorProblem    `yaml:"problem"`
	Preduplet  []Preduplet              `yaml:"preduplet"`
	Learnuplet []LearnUplet             `yaml:"learnuplet"`
}

// Storage holds metadata POSTable to storage
type Storage struct {
	Algo    []Algo    `yaml:"algo"`
	Data    []Data    `yaml:"data"`
	Model   []Model   `yaml:"model"`
	Problem []Problem `yaml:"problem"`
}

// NewDataParser parses all the metadata, and set the data folder path
func NewDataParser(pathMetadataFile, pathDataFolder string) (*DataParser, error) {
	data, err := ioutil.ReadFile(pathMetadataFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading the metadata file %s: %s", pathMetadataFile, err)
	}
	rd := &DataParser{}
	err = yaml.Unmarshal(data, rd)
	if err != nil {
		return nil, fmt.Errorf("Error Unmarshaling the metadata file %s: %s", pathMetadataFile, err)
	}
	rd.PathDataFolder = pathDataFolder
	return rd, nil
}

// GetData returns an io.Reader of the data specified
func (s *DataParser) GetData(dataType, key string) (io.ReadCloser, error) {
	path, err := FindFilePath(filepath.Join(s.PathDataFolder, dataType), key)
	if err != nil {
		return nil, fmt.Errorf("[DataParser] Error searching file %s: %s", key, err)
	}

	return os.Open(path)
}

// Print displays the data to the console
func (s *DataParser) Print() {
	fmt.Println("\n----- Orchestrator -----")
	fmt.Printf("\nAlgo: %+v\n", s.Orchestrator.Algo[0])
	fmt.Printf("\nData: %+v\n", s.Orchestrator.Data[0])
	fmt.Printf("\nPrediction: %+v\n", s.Orchestrator.Prediction[0])
	fmt.Printf("\nProblem: %+v\n", s.Orchestrator.Problem[0])
	fmt.Printf("\nLearnuplet: %+v\n", s.Orchestrator.Learnuplet[0])
	fmt.Printf("\nPreduplet: %+v\n", s.Orchestrator.Preduplet[0])

	fmt.Println("\n\n----- Storage -----")
	fmt.Printf("\nAlgo: %+v\n", s.Storage.Algo[0])
	fmt.Printf("\nData: %+v\n", s.Storage.Data[0])
	fmt.Printf("\nModel: %+v\n", s.Storage.Model[0])
	fmt.Printf("\nProblem: %+v\n", s.Storage.Problem[0])
}

// FindFilePath search recursively for a file in a folder and return its path
func FindFilePath(folder, filename string) (string, error) {
	var pathFile string
	err := filepath.Walk(folder, func(path string, f os.FileInfo, walkerr error) error {
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
