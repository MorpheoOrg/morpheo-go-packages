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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MorpheoOrg/morpheo-go-packages/common"
	"github.com/satori/go.uuid"
)

// Storage HTTP API routes
const (
	StorageProblemWorkflowRoute = "problem"
	StorageAlgoRoute            = "algo"
	StorageModelRoute           = "model"
	StorageDataRoute            = "data"
	BlobSuffix                  = "blob"
)

// Storage describes the storage service API
type Storage interface {
	GetData(id uuid.UUID) (data *common.Data, err error)
	GetAlgo(id uuid.UUID) (algo *common.Algo, err error)
	GetModel(id uuid.UUID) (model *common.Model, err error)
	GetProblemWorkflow(id uuid.UUID) (problem *common.Problem, err error)
	GetDataBlob(id uuid.UUID) (dataReader io.ReadCloser, err error)
	GetAlgoBlob(id uuid.UUID) (algoReader io.ReadCloser, err error)
	GetModelBlob(id uuid.UUID) (modelReader io.ReadCloser, err error)
	GetProblemWorkflowBlob(id uuid.UUID) (problemReader io.ReadCloser, err error)
	PostModel(model *common.Model, algoReader io.Reader, size int64) error
	PostPrediction(prediction *common.Prediction, predReader io.Reader, size int64) error
}

// StorageAPI is a wrapper around our storage HTTP API
type StorageAPI struct {
	Storage

	Hostname string
	Port     int
	User     string
	Password string
}

func (s *StorageAPI) getObjectBlob(prefix string, id uuid.UUID) (dataReader io.ReadCloser, err error) {
	url := fmt.Sprintf("http://%s:%d/%s/%s/%s", s.Hostname, s.Port, prefix, id, BlobSuffix)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("[storage-api] Error building GET request against %s: %s", url, err)
	}
	req.SetBasicAuth(s.User, s.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[storage-api] Error performing GET request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[storage-api] Bad status code (%s) performing GET request against %s", resp.Status, url)
	}

	return resp.Body, nil
}

func (s *StorageAPI) getAndParseJSONObject(objectRoute string, objectID uuid.UUID, dest interface{}) error {
	url := fmt.Sprintf("http://%s:%d/%s/%s", s.Hostname, s.Port, objectRoute, objectID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("[storage-api] Error building GET request against %s: %s", url, err)
	}
	req.SetBasicAuth(s.User, s.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[storage-api] Error performing GET request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[storage-api] Bad status code (%s) performing GET request against %s", url, err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dest)
	if err != nil {
		return fmt.Errorf("[storage-api] Error unmarshaling object retrieved from %s: %s", url, err)
	}

	return nil
}

func (s *StorageAPI) postResourceBlob(prefix string, dataReader io.Reader, size int64) error {
	url := fmt.Sprintf("http://%s:%d/%s", s.Hostname, s.Port, prefix)

	req, err := http.NewRequest(http.MethodPost, url, dataReader)
	if err != nil {
		return fmt.Errorf("[storage-api] Error building streaming POST request against %s: %s", url, err)
	}

	// Add required headers
	req.SetBasicAuth(s.User, s.Password)
	req.ContentLength = size

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[storage-api] Error performing streaming POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusCreated {
		var apiError common.APIError
		var errorMessage string
		err = json.NewDecoder(resp.Body).Decode(&apiError)
		if err != nil {
			errorMessage = "Unable to decode error message"
		}
		errorMessage = apiError.Message
		return fmt.Errorf("[storage-api] Bad status code (%s) performing streaming POST request against %s -- API Error: %s", resp.Status, url, errorMessage)
	}

	return nil
}

// postResourceMultipartBlob perform a POST request to storage using a multipart form.
// The filefield is the last field sent in the body, in order to allow streaming request.
func (s *StorageAPI) postResourceMultipartBlob(prefix string, params map[string]string, fileFieldName string, fileName string, fileReader io.Reader) error {
	// TODO: check that params are valid for the corresponding prefix

	// Build the multipart form field
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		err := writer.WriteField(key, val)
		if err != nil {
			return fmt.Errorf("Error writing param %s in %s multipart writer: %s", key, prefix, err)
		}
	}

	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		return fmt.Errorf("Error writing param blob in %s multipart writer: %s", prefix, err)
	}
	_, err = io.Copy(part, fileReader)
	if err != nil {
		return fmt.Errorf("Error copying file in %s multipart write: %s", prefix, err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("Error closing %s multipart writer: %s", prefix, err)
	}

	// Build POST request
	url := fmt.Sprintf("http://%s:%d/%s", s.Hostname, s.Port, prefix)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("[storage-api] Error building streaming POST request against %s: %s", url, err)
	}

	// Add required headers
	req.SetBasicAuth(s.User, s.Password)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform POST Request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[storage-api] Error performing streaming POST request against %s: %s", url, err)
	}

	// Handle response errors
	if resp.StatusCode != http.StatusCreated {
		var apiError common.APIError
		var errorMessage string
		err = json.NewDecoder(resp.Body).Decode(&apiError)
		if err != nil {
			errorMessage = "Unable to decode error message"
		}
		errorMessage = apiError.Message
		return fmt.Errorf("[storage-api] Bad status code (%s) performing streaming POST request against %s -- API Error: %s", resp.Status, url, errorMessage)
	}

	return nil
}

// GetProblemWorkflow returns a ProblemWorkflow's metadata
func (s *StorageAPI) GetProblemWorkflow(id uuid.UUID) (problem *common.Problem, err error) {
	problem = &common.Problem{}
	err = s.getAndParseJSONObject(StorageProblemWorkflowRoute, id, problem)
	return problem, err
}

// GetAlgo returns an Algo's metadata
func (s *StorageAPI) GetAlgo(id uuid.UUID) (algo *common.Algo, err error) {
	algo = &common.Algo{}
	err = s.getAndParseJSONObject(StorageAlgoRoute, id, algo)
	return algo, err
}

// GetModel returns a Model's metadata
func (s *StorageAPI) GetModel(id uuid.UUID) (model *common.Model, err error) {
	model = &common.Model{}
	err = s.getAndParseJSONObject(StorageModelRoute, id, model)
	return model, err
}

// GetData returns a dataset's metadata
func (s *StorageAPI) GetData(id uuid.UUID) (data *common.Data, err error) {
	data = &common.Data{}
	err = s.getAndParseJSONObject(StorageDataRoute, id, data)
	return data, err
}

// GetProblemWorkflowBlob returns an io.ReadCloser to a problem workflow image
//
// Note that it is up to the caller to call Close() on the returned io.ReadCloser
func (s *StorageAPI) GetProblemWorkflowBlob(id uuid.UUID) (dataReader io.ReadCloser, err error) {
	return s.getObjectBlob(StorageProblemWorkflowRoute, id)
}

// GetAlgoBlob returns an io.ReadCloser to a algo image (a .tar.gz file of the image's build
// context)
//
// Note that it is up to the caller to call Close() on the returned io.ReadCloser
func (s *StorageAPI) GetAlgoBlob(id uuid.UUID) (dataReader io.ReadCloser, err error) {
	return s.getObjectBlob(StorageAlgoRoute, id)
}

// GetModelBlob returns an io.ReadCloser to a model (a .tar.gz of the model volume)
//
// Note that it is up to the caller to call Close() on the returned io.ReadCloser
func (s *StorageAPI) GetModelBlob(id uuid.UUID) (dataReader io.ReadCloser, err error) {
	return s.getObjectBlob(StorageModelRoute, id)
}

// GetDataBlob returns an io.ReadCloser to a data image (a .tar.gz file of the dataset)
//
// Note that it is up to the caller to call Close() on the returned io.ReadCloser
func (s *StorageAPI) GetDataBlob(id uuid.UUID) (dataReader io.ReadCloser, err error) {
	return s.getObjectBlob(StorageDataRoute, id)
}

// PostModel returns an io.ReadCloser to a model
// TODO: change *common.Model to common.Model, and *args order
func (s *StorageAPI) PostModel(model *common.Model, modelReader io.Reader, size int64) error {
	// Check for associated Algo existence
	if _, err := s.GetAlgo(model.Algo); err != nil {
		return fmt.Errorf("Algorithm %s associated to posted model wasn't found", model.Algo)
	}

	return s.postResourceBlob(fmt.Sprintf("%s?uuid=%s&algo=%s", StorageModelRoute, model.ID, model.Algo), modelReader, size)
}

// PostProblem posts a new problem to storage
func (s *StorageAPI) PostProblem(problem common.Problem, size int, problemReader io.Reader) error {

	// Check that problem is valid
	problem.TimestampUpload = int32(time.Now().Unix())
	if err := problem.Check(); err != nil {
		return fmt.Errorf("error checking problem resource: %s", err)
	}

	// Build params
	params := make(map[string]string)
	params["uuid"] = problem.ID.String()
	params["name"] = problem.Name
	params["description"] = problem.Description
	params["size"] = strconv.Itoa(size)

	return s.postResourceMultipartBlob("problem", params, "blob", params["uuid"], problemReader)
}

// PostData posts a new data to storage
func (s *StorageAPI) PostData(data common.Data, size int, dataReader io.Reader) error {
	// Check that problem is valid
	data.TimestampUpload = int32(time.Now().Unix())
	if err := data.Check(); err != nil {
		return fmt.Errorf("error checking data resource: %s", err)
	}

	// Build params
	params := make(map[string]string)
	params["uuid"] = data.ID.String()
	params["size"] = strconv.Itoa(size)

	return s.postResourceMultipartBlob("data", params, "blob", params["uuid"], dataReader)
}

// PostPrediction posts a new prediction to storage
// TOFIX: order in PostPrediction...
func (s *StorageAPI) PostPrediction(prediction *common.Prediction, predReader io.Reader, size int64) error {
	// Check that prediction is valid
	prediction.TimestampUpload = int32(time.Now().Unix())
	if err := prediction.Check(); err != nil {
		return fmt.Errorf("error checking prediction resource: %s", err)
	}

	// Build params
	params := make(map[string]string)
	params["uuid"] = prediction.ID.String()
	params["size"] = strconv.FormatInt(size, 10)

	return s.postResourceMultipartBlob("prediction", params, "blob", params["uuid"], predReader)
}

// PostAlgo posts a new algo to storage
func (s *StorageAPI) PostAlgo(algo common.Algo, size int64, algoReader io.Reader) error {
	// Check that algo is valid
	algo.TimestampUpload = int32(time.Now().Unix())
	if err := algo.Check(); err != nil {
		return fmt.Errorf("error checking algo resource: %s", err)
	}

	// Build params
	params := make(map[string]string)
	params["uuid"] = algo.ID.String()
	params["name"] = algo.Name
	params["size"] = strconv.FormatInt(size, 10)

	return s.postResourceMultipartBlob("algo", params, "blob", params["uuid"], algoReader)
}

// StorageAPIMock is a mock of the storage API (for tests & local dev. purposes)
type StorageAPIMock struct {
	EvilUUID string
}

// NewStorageAPIMock instantiates our mock of the storage API
func NewStorageAPIMock() (*StorageAPIMock, error) {
	return &StorageAPIMock{
		EvilUUID: "610e134a-ff45-4416-aaac-1b3398e4bba6",
	}, nil
}

// GetData returns fake data (the same, no matter the UUID)
func (s *StorageAPIMock) GetData(id uuid.UUID) (*common.Data, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Data %s not found on storage", id)
	}

	return common.NewData(), nil
}

// GetAlgo returns a fake algo, no matter the UUID
func (s *StorageAPIMock) GetAlgo(id uuid.UUID) (*common.Algo, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Algo %s not found on storage", id)
	}

	return common.NewAlgo(), nil
}

// GetModel returns a fake model, no matter the UUID
func (s *StorageAPIMock) GetModel(id uuid.UUID) (*common.Model, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Model %s not found on storage", id)
	}
	algo := common.NewAlgo()
	return common.NewModel(id, algo), nil
}

// GetProblemWorkflow returns a fake algo, no matter the UUID
func (s *StorageAPIMock) GetProblemWorkflow(id uuid.UUID) (*common.Problem, error) {
	// Evil uuid returns Error
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Problem workflow %s not found on storage", id)
	}
	return common.NewProblem(), nil
}

// GetDataBlob returns a fake Data, no matter the UUID
func (s *StorageAPIMock) GetDataBlob(id uuid.UUID) (io.ReadCloser, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Data blob %s not found on storage", id)
	}

	return TargzedMock()
}

// GetAlgoBlob returns a fake Algo, no matter the UUID
func (s *StorageAPIMock) GetAlgoBlob(id uuid.UUID) (io.ReadCloser, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Algo blob %s not found on storage", id)
	}
	return TargzedMock()
}

// GetModelBlob returns a fake Model, no matter the UUID
func (s *StorageAPIMock) GetModelBlob(id uuid.UUID) (io.ReadCloser, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("Model blob %s not found on storage", id)
	}
	return TargzedMock()
}

// GetProblemWorkflowBlob returns a fake ProblemWorkflow, no matter the UUID
func (s *StorageAPIMock) GetProblemWorkflowBlob(id uuid.UUID) (io.ReadCloser, error) {
	if id.String() == s.EvilUUID {
		return nil, fmt.Errorf("ProblemWorkflow blob %s not found on storage", id)
	}
	return TargzedMock()
}

// PostModel sends a model... to Oblivion
func (s *StorageAPIMock) PostModel(model *common.Model, modelReader io.Reader, size int64) error {
	_, err := io.Copy(ioutil.Discard, modelReader)
	return err
}

// PostPrediction sends a prediction... to Oblivion
func (s *StorageAPIMock) PostPrediction(prediction *common.Prediction, predReader io.Reader, size int64) error {
	_, err := io.Copy(ioutil.Discard, predReader)
	return err
}

// TargzedMock create a Readcloser which can be ungzip-ed
func TargzedMock() (io.ReadCloser, error) {
	// Create tmp file
	tmpPath := filepath.Join(os.TempDir(), "morpheo_mock")
	if err := ioutil.WriteFile(tmpPath, []byte("mock"), 0777); err != nil {
		return nil, fmt.Errorf("Error writing file: %s", err)
	}
	f, _ := os.Open(tmpPath)
	defer os.Remove(tmpPath)

	buf := bytes.NewBuffer([]byte(""))
	if err := TargzFile(f, buf); err != nil {
		return nil, fmt.Errorf("Error Targz-ing the file: %s", err)
	}

	return ioutil.NopCloser(buf), nil
}

// TargzFile tars and gzips a file and forwards it to an io.Writer
func TargzFile(file *os.File, dest io.Writer) error {
	// Let's wire our writer together
	zipWriter := gzip.NewWriter(dest)
	defer zipWriter.Close()
	tarWriter := tar.NewWriter(zipWriter)
	defer tarWriter.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Error getting file info: %s", err)
	}
	// Let's create the header
	header := &tar.Header{
		Name:    stat.Name(),
		Size:    stat.Size(),
		Mode:    0664,
		ModTime: stat.ModTime(),
	}
	// write the header to the tarball archive
	if err = tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("Error writing tar header for file %s", err)
	}
	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("Error writing file %s to tar archive", err)
	}
	return nil
}
