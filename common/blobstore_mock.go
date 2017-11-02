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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// This little perverts make Put and Get calls fail
const (
	NaughtySize      = 69
	ViciousDevilUUID = "2cd41d08-ef54-4a15-95a1-2e84ca72a22c"
)

// MOCKBlobStore is a BlobStore implementations for tests
type MOCKBlobStore struct {
}

// NewMOCKBlobStore creates a new Blobstore for tests
func NewMOCKBlobStore(dataDir string) (BlobStore, error) {
	if dataDir == "evil" {
		return nil, fmt.Errorf("[fake-blobstore] Evil blobStore")
	}
	return &MOCKBlobStore{}, nil
}

// Put writes a file in the data directory (and creates necessarry sub-directories if there are
// forward slashes in the key name)
func (s *MOCKBlobStore) Put(key string, data io.Reader, size int64) error {
	if size == NaughtySize {
		return fmt.Errorf("[fake-blobstore] What a naughty size")
	}
	return nil
}

// Get returns an io.ReadCloser on the data living under the provided key. The retriever must
// explicitely call the Close() method on it when he's done reading.
func (s *MOCKBlobStore) Get(key string) (data io.ReadCloser, err error) {
	// Check if uuid (end of key) is the ViciousDevilUUID
	if strings.SplitAfter(key, "/")[1] == ViciousDevilUUID {
		return nil, fmt.Errorf("[fake-blobstore] Runnin' With the Devil")
	}
	return fakeFile(), nil
}

// Delete remove the file
func (s *MOCKBlobStore) Delete(key string) (err error) {
	return nil
}

// Rename renames the file
func (s *MOCKBlobStore) Rename(key string, newKey string) (err error) {
	return nil
}

func fakeFile() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer([]byte("fakeFileContent")))
}
