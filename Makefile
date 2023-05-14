# Copyright 2023 Codenotary Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# 	http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO ?= go
V_VERSION ?= devbuild
V_COMMIT ?= $(shell git rev-parse HEAD|head -c 7)
V_BUILT_AT = $(shell date +%s)
BUILD_TAG ?= $(shell git rev-parse HEAD|head -c 8)

V_LDFLAGS = -X "github.com/codenotary/immudb-log-audit/cmd.Version=$(V_VERSION)"\
	-X "github.com/codenotary/immudb-log-audit/cmd.BuildTime=$(V_BUILT_AT)"\
	-X "github.com/codenotary/immudb-log-audit/cmd.Commit=$(V_COMMIT)"\

REPO ?= codenotary
IMAGE_TAG ?= $(REPO)/immudb-log-audit
PUSH_TAG ?= $(REPO)/immudb-log-audit

.PHONY: immudb-log-audit
immudb-log-audit:
	$(GO) build -v -ldflags '$(V_LDFLAGS)' -o immudb-log-audit ./main.go

.PHONY: test
test:
	$(GO) vet ./...
	$(GO) get golang.org/x/tools/cmd/cover
	$(GO) get github.com/mattn/goveralls
	$(GO) get -u github.com/ory/go-acc
	go-acc ./... --ignore test || true
	$(GO) tool cover -func=coverage.txt | grep total | grep -Eo '[0-9]+\.[0-9]+'

.PHONY: docker push
docker:
	docker build . -t $(IMAGE_TAG) \
	--label "com.codenotary.commit=$(BUILD_TAG)" \
	--build-arg V_VERSION=$(V_VERSION) \
	--build-arg V_COMMIT=$(V_COMMIT) \
	--build-arg V_BUILT_AT=$(V_BUILT_AT) 

push: docker
	docker image tag $(IMAGE_TAG):latest $(PUSH_TAG):$(BUILD_TAG)
	docker image tag $(IMAGE_TAG):latest $(PUSH_TAG):latest
	docker image push $(PUSH_TAG):$(BUILD_TAG)
	docker image push $(PUSH_TAG):latest