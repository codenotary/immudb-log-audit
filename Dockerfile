# Copyright 2023 Codenotary Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# 	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.19-alpine as build
WORKDIR /src
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -v

FROM scratch
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /src/immudb-log-audit /app/immudb-log-audit

ENTRYPOINT [ "/app/immudb-log-audit" ]