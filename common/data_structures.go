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
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

// Uplet types
const (
	TypeLearnuplet = "learnuplet"
	TypePredUplet  = "preduplet"
)

var (
	// ValidUplets us a set of all possible uplet names
	ValidUplets = map[string]struct{}{
		TypeLearnuplet: struct{}{},
		TypePredUplet:  struct{}{},
	}
)

// Task statuses
const (
	TaskStatusTodo    = "todo"
	TaskStatusPending = "pending"
	TaskStatusDone    = "done"
	TaskStatusFailed  = "failed"
)

var (
	// ValidStatuses is a set of all possible values for the "status" field
	ValidStatuses = map[string]struct{}{
		TaskStatusTodo:    struct{}{},
		TaskStatusPending: struct{}{},
		TaskStatusDone:    struct{}{},
		TaskStatusFailed:  struct{}{},
	}
)

// ===========================================================================
// Chaincode Data Structures: LearnupletChaincode
// ===========================================================================
// Used by
// ====================================
//   - Compute-api to QUERY and convert learnuplets items from Peer Client
//
//  N.B.:
//   - NOT USED by Peer Client as return values to Learnuplet queries
//   - NOT USED by Orchestrator chaincode
//
// Functions
// ====================================
// LearnupletFormat: Convert LearnupletChaincode into Learnuplet
// GetUUIDFromKey:   Convert <object_uuid> into <uuid>

// LearnupletChaincode describes a QUERY (queryItem, learnuplet) to the chaincode
type LearnupletChaincode struct {
	Key                   string             `json:"key"`
	ProblemStorageAddress string             `json:"problem_storage_address"`
	Algo                  string             `json:"algo"`
	ModelStart            string             `json:"model_start"`
	ModelEnd              string             `json:"model_end"`
	TrainData             []string           `json:"train_data"`
	TestData              []string           `json:"test_data"`
	Worker                string             `json:"worker"`
	Status                string             `json:"status"`
	Rank                  int                `json:"rank"`
	Perf                  float64            `json:"perf"`
	TrainPerf             map[string]float64 `json:"train_perf"`
	TestPerf              map[string]float64 `json:"test_perf"`
}

// LearnupletFormat convert LearnupletChaincode into Learnuplet
func (s *LearnupletChaincode) LearnupletFormat() (Learnuplet, error) {
	problem, err := uuid.FromString(s.ProblemStorageAddress)
	if err != nil {
		return Learnuplet{}, fmt.Errorf("Failed to parse ProblemStorageAddress UUID: %s", err)
	}

	algo, err := GetUUIDFromKey(s.Algo)
	if err != nil {
		return Learnuplet{}, fmt.Errorf("Failed to parse Algo UUID: %s", err)
	}

	var modelStart uuid.UUID
	if s.ModelStart != "" {
		modelStart, err = uuid.FromString(s.ModelStart)
		if err != nil {
			return Learnuplet{}, fmt.Errorf("Failed to parse ModelStart UUID: %s", err)
		}
	}

	var modelEnd uuid.UUID
	if s.ModelEnd != "" {
		modelEnd, err = uuid.FromString(s.ModelEnd)
		if err != nil {
			return Learnuplet{}, fmt.Errorf("Failed to parse ModelEnd UUID: %s", err)
		}
	}

	var trainData []uuid.UUID
	for _, data := range s.TrainData {
		dataID, err := GetUUIDFromKey(data)
		if err != nil {
			return Learnuplet{}, fmt.Errorf("Failed to parse TrainData UUID: %s", err)
		}
		trainData = append(trainData, dataID)
	}

	var testData []uuid.UUID
	for _, data := range s.TestData {
		dataID, err := GetUUIDFromKey(data)
		if err != nil {
			return Learnuplet{}, fmt.Errorf("Failed to parse TestData UUID: %s", err)
		}
		testData = append(testData, dataID)
	}

	var worker uuid.UUID
	if s.Worker == "" {
		worker = uuid.Nil
	} else {
		worker, err = uuid.FromString(s.Worker)
		if err != nil {
			return Learnuplet{}, fmt.Errorf("Failed to parse Worker UUID: %s", err)
		}
	}
	return Learnuplet{
		Key:         s.Key,
		Problem:     problem,
		TrainData:   trainData,
		TestData:    testData,
		Algo:        algo,
		ModelStart:  modelStart,
		ModelEnd:    modelEnd,
		Rank:        s.Rank,
		Worker:      worker,
		Status:      s.Status,
		RequestDate: int(time.Now().Unix()),
	}, nil
}

// GetUUIDFromKey returns the uuid of a given key
func GetUUIDFromKey(key string) (uuid.UUID, error) {
	keySplit := strings.Split(key, "_")
	if len(keySplit) != 2 {
		return uuid.Nil, fmt.Errorf("Wrong format for Key: should be <object_uuid>, have: \"%s\"", key)
	}
	id, err := uuid.FromString(keySplit[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("Failed to parse UUID %s", keySplit[1])
	}
	return id, nil
}

// ===========================================================================
// Compute Data Structures
// ===========================================================================
// The Resources used by Storage are described here: Algo, Data, Model,
// Prediction and Problem.
//
// Used by
// ====================================
//   - Compute-api to Push learnuplets to the broker
//   - Compute-worker to retrieve a learning task from broker
//
//  N.B.: NOT USED by Compute-api to query Chaincode (see LearnupletChaincode)
//
// Functions
// ====================================
// Check: Check that the resource struct fields are correctly set

// Checkable is an Interface for things that can be Checked (i.e. validated after a JSON parsing for
// instance)
type Checkable interface {
	Check() (err error)
}

// Learnuplet describes a Learning task.
type Learnuplet struct {
	Key            string      `json:"key" yaml:"key"`
	Problem        uuid.UUID   `json:"problem" yaml:"problem"`
	TrainData      []uuid.UUID `json:"train_data" yaml:"train_data"`
	TestData       []uuid.UUID `json:"test_data" yaml:"test_data"`
	Algo           uuid.UUID   `json:"algo" yaml:"algo"`
	ModelStart     uuid.UUID   `json:"model_start" yaml:"model_start"`
	ModelEnd       uuid.UUID   `json:"model_end" yaml:"model_end"`
	Rank           int         `json:"rank" yaml:"rank"`
	Worker         uuid.UUID   `json:"worker" yaml:"worker"` // @camillemarini: I didn't get the purpose of this field
	Status         string      `json:"status" yaml:"status"`
	RequestDate    int         `json:"timestamp_request" yaml:"timestamp_request"`
	CompletionDate int         `json:"timestamp_done" yaml:"timestamp_done"`
}

// Preduplet describes a prediction task.
type Preduplet struct {
	ID                  uuid.UUID `json:"uuid" yaml:"uuid"`
	Problem             uuid.UUID `json:"problem" yaml:"problem"`
	Model               uuid.UUID `json:"model" yaml:"model"`
	Data                uuid.UUID `json:"data" yaml:"data"`
	Worker              uuid.UUID `json:"worker" yaml:"worker"`
	Status              string    `json:"status" yaml:"status"`
	RequestDate         int       `json:"timestamp_request" yaml:"timestamp_request"`
	CompletionDate      int       `json:"timestamp_done" yaml:"timestamp_done"`
	PredictionStorageID uuid.UUID `json:"prediction_storage_uuid" yaml:"prediction_storage_uuid"`
}

// Compute Specific Functions: Check
// ===========================================================================

// Check returns nil if the learnuplet is valid, an explicit error otherwise
func (s *Learnuplet) Check() (err error) {

	if s.Key == "" {
		return fmt.Errorf("id field is required")
	}

	if uuid.Equal(uuid.Nil, s.Problem) {
		return fmt.Errorf("problem field is required")
	}

	if uuid.Equal(uuid.Nil, s.Algo) {
		return fmt.Errorf("algo field is required")
	}

	if len(s.TrainData) == 0 {
		return fmt.Errorf("train_data field is empty or unset")
	}
	for n, id := range s.TrainData {
		if uuid.Equal(uuid.Nil, id) {
			return fmt.Errorf("Nil UUID in train_data field at pos %d", n)
		}
	}

	if len(s.TestData) == 0 {
		return fmt.Errorf("test_data field is empty or unset")
	}
	for n, id := range s.TestData {
		if uuid.Equal(uuid.Nil, id) {
			return fmt.Errorf("Empty UUID in test_data field at pos %d", n)
		}
	}

	if _, ok := ValidStatuses[s.Status]; !ok {
		return fmt.Errorf("status field ain't valid (provided: %s, possible choices: %s", s.Status, ValidStatuses)
	}

	if s.Rank > 0 {
		if uuid.Equal(uuid.Nil, s.ModelStart) {
			return fmt.Errorf("rank %d and empty ModelStart", s.Rank)
		}
	}

	return nil
}

// Check returns nil if the preduplet is valid, an explicit error otherwise
func (s *Preduplet) Check() (err error) {

	if uuid.Equal(uuid.Nil, s.ID) {
		return fmt.Errorf("id field is unset")
	}
	if uuid.Equal(uuid.Nil, s.Problem) {
		return fmt.Errorf("problem field is unset")
	}
	if uuid.Equal(uuid.Nil, s.Model) {
		return fmt.Errorf("model field is required")
	}
	if len(s.Data) == 0 {
		return fmt.Errorf("data field is empty or unset")
	}
	if uuid.Equal(uuid.Nil, s.Data) {
		return fmt.Errorf("Nil UUID in data field")
	}
	if _, ok := ValidStatuses[s.Status]; !ok {
		return fmt.Errorf("status field ain't valid (provided: %s, possible choices: %s", s.Status, ValidStatuses)
	}

	return nil
}

// ===========================================================================
// Storage Data Structures
// ===========================================================================
// The Resources used by Storage are described here: Algo, Data, Model,
// Prediction and Problem.
//
// Used by
// ====================================
//   - Storage to parse POST request
//   - Go-package/client to post data on Storage
//
// Functions
// ====================================
// New<Resource>: Create a new data structure <Algo/Data/Model/Prediction/Problem>
// Check:         Check that the resource struct fields are correctly set
// GetUUID:       Retrieve the ID of the Resource
// FillResource:  Fill a resource with the elements in a Go map

// Resource is an interface for Storage Resources
type Resource interface {
	GetUUID() uuid.UUID
	FillResource(fields map[string]interface{}) error
	Check() error
}

// Algo describes an algorithm blob
type Algo struct {
	ID              uuid.UUID `json:"uuid" yaml:"uuid" db:"uuid"`
	TimestampUpload int32     `json:"timestamp_upload" yaml:"timestamp_upload" db:"timestamp_upload"`
	Name            string    `json:"name" yaml:"name" db:"name"`
}

// Data describes a data blob
type Data struct {
	ID              uuid.UUID `json:"uuid" yaml:"uuid" db:"uuid"`
	TimestampUpload int32     `json:"timestamp_upload" yaml:"timestamp_upload" db:"timestamp_upload"`
}

// Model describes a model blob (should be a .tar.gz of the model folder)
type Model struct {
	ID              uuid.UUID `json:"uuid" yaml:"uuid" db:"uuid"`
	TimestampUpload int32     `json:"timestamp_upload" yaml:"timestamp_upload" db:"timestamp_upload"`
	Algo            uuid.UUID `json:"algo" yaml:"algo" db:"algo"`
}

// Prediction describes a prediction blob
type Prediction struct {
	ID              uuid.UUID `json:"uuid" yaml:"uuid" db:"uuid"`
	TimestampUpload int32     `json:"timestamp_upload" yaml:"timestamp_upload" db:"timestamp_upload"`
}

// Problem describes a problem blob
type Problem struct {
	ID              uuid.UUID `json:"uuid" yaml:"uuid" db:"uuid"`
	TimestampUpload int32     `json:"timestamp_upload" yaml:"timestamp_upload" db:"timestamp_upload"`
	Name            string    `json:"name" yaml:"name" db:"name"`
	Description     string    `json:"description" yaml:"description" db:"description"`
}

// Functions to create new data structures
// ===========================================================================

// NewAlgo creates an Algo instance
func NewAlgo() *Algo {
	algo := &Algo{
		ID:              uuid.NewV4(),
		TimestampUpload: int32(time.Now().Unix()),
	}
	return algo
}

// NewData creates a problem instance
func NewData() *Data {
	data := &Data{
		ID:              uuid.NewV4(),
		TimestampUpload: int32(time.Now().Unix()),
	}
	return data
}

// NewModel creates a model instance - Used by Storage AND Compute
func NewModel(id uuid.UUID, algo *Algo) *Model {
	idModel := id
	if idModel == uuid.Nil {
		idModel = uuid.NewV4()
	}
	model := &Model{
		ID:              idModel,
		TimestampUpload: int32(time.Now().Unix()),
		Algo:            algo.ID,
	}
	return model
}

// NewPrediction creates a prediction instance
func NewPrediction() *Prediction {
	prediction := &Prediction{
		ID:              uuid.NewV4(),
		TimestampUpload: int32(time.Now().Unix()),
	}
	return prediction
}

// NewProblem creates a problem instance
func NewProblem() *Problem {
	problem := &Problem{
		ID:              uuid.NewV4(),
		TimestampUpload: int32(time.Now().Unix()),
	}
	return problem
}

// Storage Specific Functions: Check, GetUUID and FillResource
// ===========================================================================
// ALGO

// Check returns nil if the Resource is correctly filled
func (a *Algo) Check() error {
	if uuid.Equal(uuid.Nil, a.ID) {
		return fmt.Errorf("'UUID' unset")
	}
	if a.Name == "" {
		return fmt.Errorf("'Name' unset")
	}
	if a.TimestampUpload <= 0 {
		return fmt.Errorf("'Timestamp_upload' unset")
	}
	return nil
}

// GetUUID returns the resource uuid
func (a *Algo) GetUUID() uuid.UUID {
	return a.ID
}

// FillResource fills the resource with elements in a map
func (a *Algo) FillResource(fields map[string]interface{}) error {
	for k, v := range fields {
		switch k {
		case "uuid":
			a.ID = v.(uuid.UUID)
		case "name":
			a.Name = v.(string)
		default:
			return fmt.Errorf("%s is not a valid field for algo", k)
		}
	}
	a.TimestampUpload = int32(time.Now().Unix())
	return nil
}

// DATA

// Check returns nil if the Resource is correctly filled
func (d *Data) Check() error {
	if uuid.Equal(uuid.Nil, d.ID) {
		return fmt.Errorf("'UUID' unset")
	}
	if d.TimestampUpload <= 0 {
		return fmt.Errorf("'Timestamp_upload' unset")
	}
	return nil
}

// GetUUID returns the resource uuid
func (d *Data) GetUUID() uuid.UUID {
	return d.ID
}

// FillResource fills the resource with elements in a map
func (d *Data) FillResource(fields map[string]interface{}) error {
	// TODO: Try generic func with reflection
	for k, v := range fields {
		switch k {
		case "uuid":
			d.ID = v.(uuid.UUID) // TODO: handle errors with type assertion...
		default:
			return fmt.Errorf("%s is not a valid field for data", k)
		}
	}
	d.TimestampUpload = int32(time.Now().Unix())
	return nil
}

// PREDICTION

// Check returns nil if the Resource is correctly filled
func (p *Prediction) Check() error {
	if uuid.Equal(uuid.Nil, p.ID) {
		return fmt.Errorf("'UUID' unset")
	}
	if p.TimestampUpload <= 0 {
		return fmt.Errorf("'Timestamp_upload' unset")
	}
	return nil
}

// GetUUID returns the resource uuid
func (p *Prediction) GetUUID() uuid.UUID {
	return p.ID
}

// FillResource fills the resource with elements in a map
func (p *Prediction) FillResource(fields map[string]interface{}) error {
	for k, v := range fields {
		switch k {
		case "uuid":
			p.ID = v.(uuid.UUID) // TODO: handle errors with type assertion...
		default:
			return fmt.Errorf("%s is not a valid field for problem", k)
		}
	}
	p.TimestampUpload = int32(time.Now().Unix())
	return nil
}

// PROBLEM

// Check returns nil if the Resource is correctly filled
func (p *Problem) Check() error {
	if uuid.Equal(uuid.Nil, p.ID) {
		return fmt.Errorf("'UUID' unset")
	}
	if p.Name == "" {
		return fmt.Errorf("'Name' unset")
	}
	if p.Description == "" {
		return fmt.Errorf("'Description' unset")
	}
	if p.TimestampUpload <= 0 {
		return fmt.Errorf("'Timestamp_upload' unset")
	}
	return nil
}

// GetUUID returns the resource uuid
func (p *Problem) GetUUID() uuid.UUID {
	return p.ID
}

// FillResource fills the resource with elements in a map
func (p *Problem) FillResource(fields map[string]interface{}) error {
	// TODO: Try generic func with reflection
	for k, v := range fields {
		switch k {
		case "uuid":
			p.ID = v.(uuid.UUID) // TODO: handle errors with type assertion...
		case "name":
			p.Name = v.(string)
		case "description":
			p.Description = v.(string)
		default:
			return fmt.Errorf("%s is not a valid field for problem", k)
		}
	}
	p.TimestampUpload = int32(time.Now().Unix())
	return nil
}

// ===========================================================================
// Errors management
// ===========================================================================
// Used by
// ====================================
//   - Compute-api
//   - Storage
//   - Go-packages/client Storage

// APIError wraps errors sent back by the HTTP API
type APIError struct {
	Message string `json:"error"`
	Status  int    `json:"status"`
}

// NewAPIError creates an APIError object, given an error message
func NewAPIError(message string) (err *APIError) {
	return &APIError{
		Message: message,
	}
}

// Error returns the error message as a string
func (err *APIError) Error() string {
	return err.Message
}

// TaskError describes an error happening in the consumer that indicates the errord task can be
// retried (if the retry limit hasn't been reached)
type TaskError struct {
	Message string `json:"error"`
}

func (e *TaskError) Error() string {
	return e.Message
}

// FatalTaskError describes an error happening in the consumer that isn't worth a retry
type FatalTaskError struct {
	Message string `json:"error"`
}

func (e *FatalTaskError) Error() string {
	return e.Message
}
