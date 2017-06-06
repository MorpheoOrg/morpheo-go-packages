#
# Copyright Morpheo Org. 2017
#
# contact@morpheo.co
#
# This software is part of the Morpheo project, an open-source machine
# learning platform.
#
# This software is governed by the CeCILL license, compatible with the
# GNU GPL, under French law and abiding by the rules of distribution of
# free software. You can  use, modify and/ or redistribute the software
# under the terms of the CeCILL license as circulated by CEA, CNRS and
# INRIA at the following URL "http://www.cecill.info".
#
# As a counterpart to the access to the source code and  rights to copy,
# modify and redistribute granted by the license, users are provided only
# with a limited warranty  and the software's author,  the holder of the
# economic rights,  and the successive licensors  have only  limited
# liability.
#
# In this respect, the user's attention is drawn to the risks associated
# with loading,  using,  modifying and/or developing or reproducing the
# software by the user in light of its specific status of free software,
# that may mean  that it is complicated to manipulate,  and  that  also
# therefore means  that it is reserved for developers  and  experienced
# professionals having in-depth computer knowledge. Users are therefore
# encouraged to load and test the software's suitability as regards their
# requirements in conditions enabling the security of their systems and/or
# data to be ensured and,  more generally, to use and operate it in the
# same conditions as regards security.
#
# The fact that you are presently reading this means that you have had
# knowledge of the CeCILL license and that you accept its terms.
#

# (Containerized) build commands
BUILD_CONTAINER = \
  docker run -u $(shell id -u) -it --rm \
	  --workdir "/usr/local/go/src/github.com/MorpheoOrg/go-packages" \
	  -v $${PWD}:/usr/local/go/src/github.com/MorpheoOrg/go-packages:ro \
	  -v $${PWD}/vendor:/vendor/src \
	  -e GOPATH="/go:/vendor" \
	  -e CGO_ENABLED=0 \
	  -e GOOS=linux

GLIDE_CONTAINER = \
	docker run -it --rm \
	  --workdir "/usr/local/go/src/github.com/MorpheoOrg/go-morpheo" \
	  -v $${PWD}:/usr/local/go/src/github.com/MorpheoOrg/go-morpheo \
		$(BUILD_CONTAINER_IMAGE)

BUILD_CONTAINER_IMAGE = golang:1-onbuild

GOBUILD = go build --installsuffix cgo --ldflags '-extldflags \"-static\"'
GOTEST = go test

# Targets (files & phony targets)
TARGETS = client common
TEST_TARGETS = $(foreach TARGET,$(TARGETS),$(TARGET)-test)

# Target configuration
.DEFAULT: all
.PHONY: all clean vendor-clean test $(TEST_TARGETS)

# Project wide targets
test: $(TEST_TARGETS)
clean: vendor-clean

# 1. Vendoring
vendor: glide.yaml
	@echo "Pulling dependencies with glide... in a build container too"
	rm -rf ./vendor
	mkdir ./vendor
	$(GLIDE_CONTAINER) bash -c \
		"go get github.com/Masterminds/glide && glide install && chown $(shell id -u):$(shell id -g) -R ./glide.lock ./vendor"

vendor-clean:
	@echo "Dropping the vendor folder"
	rm -rf ./vendor

# 2. Testing
$(TEST_TARGETS): vendor
	@echo "Running go test in $(subst -test,,$(@)) directory"
	$(BUILD_CONTAINER) $(BUILD_CONTAINER_IMAGE) \
    bash -c "cd $(subst -test,,$(@)) && $(GOTEST) "
