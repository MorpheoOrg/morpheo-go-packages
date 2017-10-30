package common

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	// TODO: relative path
	PathFixtures         = filepath.Join(os.Getenv("GOPATH"), "src/github.com/MorpheoOrg/morpheo-devenv/tests/fixtures.yaml")
	RootPathFixturesData = filepath.Join(os.Getenv("GOPATH"), "src/github.com/MorpheoOrg/morpheo-devenv/data/fixtures")
)

var (
	PathFixturesData = map[string]string{
		"algo":    filepath.Join(RootPathFixturesData, "algo/example_sklearn"),
		"data":    filepath.Join(RootPathFixturesData, "data/test"),
		"problem": filepath.Join(RootPathFixturesData, "problem/mesa"),
		"model":   filepath.Join(RootPathFixturesData, "model"),
		"perf":    filepath.Join(RootPathFixturesData, "perf"),
		"pred":    filepath.Join(RootPathFixturesData, "pred"),
	}
)

type Fixtures struct {
	Orchestrator Orchestrator `yaml:"orchestrator"`
}

type Orchestrator struct {
	Preduplet  []Preduplet  `yaml:"preduplet"`
	Learnuplet []LearnUplet `yaml:"learnuplet"`
}

func LoadFixtures() (*Fixtures, error) {
	data, err := ioutil.ReadFile(PathFixtures)
	if err != nil {
		return nil, fmt.Errorf("Error reading file fixtures.yaml: %s", err)
	}
	f := &Fixtures{}
	err = yaml.Unmarshal(data, f)
	if err != nil {
		return nil, fmt.Errorf("Error Unmarshaling fixtures.yaml file: %s", err)
	}
	return f, nil
}
