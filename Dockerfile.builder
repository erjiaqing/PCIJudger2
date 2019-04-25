FROM golang:1.12 as file_container

COPY /lang /fj/lang
COPY /kotlinc /fj/kotlinc
COPY /lrun /fj/lrun
COPY /support /fj/assets
COPY ["mirrorfs.conf", "/fj/"]

FROM golang:1.12 as builder
COPY /vendor /go/src/github.com/erjiaqing/PCIJudger2/vendor
COPY /cmd /go/src/github.com/erjiaqing/PCIJudger2/cmd
COPY /pkg /go/src/github.com/erjiaqing/PCIJudger2/pkg
RUN cd /go/src/github.com/erjiaqing/PCIJudger2/cmd/pci15/judger && go build && cd /go/src/github.com/erjiaqing/PCIJudger2/cmd/pci15/builder && go build


FROM ubuntu:18.04
VOLUME ["/problem", "/code"]

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install software-properties-common -y
RUN echo 'deb https://mirrors.tuna.tsinghua.edu.cn/ubuntu/ bionic main restricted universe multiverse\ndeb https://mirrors.tuna.tsinghua.edu.cn/ubuntu/ bionic-updates main restricted universe multiverse\ndeb https://mirrors.tuna.tsinghua.edu.cn/ubuntu/ bionic-backports main restricted universe multiverse\ndeb https://mirrors.tuna.tsinghua.edu.cn/ubuntu/ bionic-security main restricted universe multiverse\n' > /etc/apt/sources.list && \
    find /etc/apt/sources.list.d/ -type f -name "*.list" -exec  sed  -i.bak -r  's#deb(-src)?\s*http(s)?://ppa.launchpad.net#deb\1 http\2://launchpad.proxy.ustclug.org#ig' {} \;
RUN add-apt-repository ppa:longsleep/golang-backports && \
    apt-get update && \
    apt-get install -y \
        build-essential \
        python3 \
        golang-1.11 \
        mono-mcs mono-runtime \
        openjdk-8-jdk-headless \
        fpc \
        php-cli \
        libseccomp-dev \
        rake \
        ghc && \
    rm -rf /var/lib/apt/lists/* && \
    apt clean
COPY --from=file_container /fj /fj
RUN cd /fj/lrun && make install && make clean && useradd runner && adduser runner lrun
COPY --from=builder /go/src/github.com/erjiaqing/PCIJudger2/cmd/pci15/builder/builder /fj/builder

WORKDIR /fj/
USER runner
ENTRYPOINT ["/fj/builder", "-input", "/problem", "-langconf", "/fj/lang", "-assets", "/fj/assets"]
