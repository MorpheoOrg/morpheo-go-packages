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
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

// GCBlobStore implements the interface Blobstore for Google Cloud Storage
type GCBlobStore struct {
	bucket *storage.BucketHandle
}

// NewGCBlobStore creates a new Google Cloud Blobstore
func NewGCBlobStore(bucketName string) (*GCBlobStore, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("[gc-storage] Error creating client: %s", err)
	}
	bucket := client.Bucket(bucketName)
	return &GCBlobStore{
		bucket: bucket,
	}, nil
}

// Put streams a file to GC
func (s *GCBlobStore) Put(key string, r io.Reader, size int64) error {
	obj := s.bucket.Object(key)

	w := obj.NewWriter(context.Background())

	n, err := io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("[gc-storage] Error uploading file (%d bytes written): %s", n, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("[gc-storage] Error uploading file: Error closing object: %s", err)
	}
	return nil
}

// Get retrieve a data with the specified uuid
func (s *GCBlobStore) Get(key string) (io.ReadCloser, error) {
	obj := s.bucket.Object(key)
	r, err := obj.NewReader(context.Background())
	if err != nil {
		return nil, fmt.Errorf("[gc-storage] Error retrieving file: %s", err)
	}
	return r, nil
}

// Delete deletes a data with the specified uuid
func (s *GCBlobStore) Delete(key string) error {
	obj := s.bucket.Object(key)
	if err := obj.Delete(context.Background()); err != nil {
		return fmt.Errorf("[gc-storage] Error deleting file: %s", err)
	}
	return nil
}

// Rename renames a data with the specified uuid
func (s *GCBlobStore) Rename(key string, newKey string) error {
	objSrc := s.bucket.Object(key)
	objDest := s.bucket.Object(newKey)
	if _, err := objDest.CopierFrom(objSrc).Run(context.Background()); err != nil {
		return fmt.Errorf("[gc-storage] Error renaming file: %s", err)
	}
	if err := objSrc.Delete(context.Background()); err != nil {
		return fmt.Errorf("[gc-storage] Error deleting old file: %s", err)
	}
	return nil
}
