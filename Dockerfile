# Copyright 2022 Linka Cloud  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.20 as builder

WORKDIR /d2vm

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN make .build

FROM ubuntu:20.04

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends \
        ca-certificates \
        util-linux \
        udev \
        parted \
        kpartx \
        e2fsprogs \
        mount \
        tar \
        extlinux \
        cryptsetup-bin \
        qemu-utils && \
    apt-get -y install --no-install-recommends openssh-server && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -rm -d /home/ubuntu -s /bin/bash -g root -G sudo -u 1000 test

RUN  echo 'test:test' | chpasswd

RUN service ssh start

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]

COPY --from=docker:dind /usr/local/bin/docker /usr/local/bin/

COPY --from=builder /d2vm/d2vm /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/d2vm"]
