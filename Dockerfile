FROM golang:1.11 as file_container

COPY /lang /fj/lang
COPY /kotlinc /fj/kotlinc
COPY /lrun /fj/lrun

FROM golang:1.11 as builder
COPY /vendor /go/src/github.com/erjiaqing/PCIJudger2/vendor
COPY /cmd /go/src/github.com/erjiaqing/PCIJudger2/cmd
COPY /pkg /go/src/github.com/erjiaqing/PCIJudger2/pkg
COPY ["mirrorfs.conf", "/fj/"]
RUN cd /go/src/github.com/erjiaqing/PCIJudger2/cmd/pci15/judger && go build


FROM ubuntu:16.04
VOLUME ["/problem", "/code"]

RUN apt-get update && apt-get install software-properties-common -y && \
    add-apt-repository ppa:gophers/archive && \
    apt-get update && \
    apt-get install -y \
        build-essential \
        python3 python3-pip \
        golang-1.9-go \
        mono-mcs mono-runtime \
        openjdk-8-jdk-headless \
        fpc \
        php7.0-cli \
        libseccomp-dev \
        rake \
        ghc && \
    rm -rf /var/lib/apt/lists/* && \
    apt clean && \
    pip3 install PyYAML
COPY --from=file_container /fj /fj
RUN cd /fj/lrun && make install && make clean && useradd runner && adduser runner lrun
COPY --from=builder /go/src/github.com/erjiaqing/PCIJudger2/cmd/pci15/judger/judger /fj/judger

WORKDIR /fj/
USER runner
ENTRYPOINT ["/fj/judger", "-problem", "/problem", "-langconf", "/fj/lang", "-mirrorfsconf", "/fj/mirrorfs.conf", "-source", "/code/code", "-thread", "4"]