FROM golang:1.15 AS build
WORKDIR /app
RUN git clone https://github.com/Zilliqa/zilliqa-relayer.git  && \
    cd zilliqa-relayer && \
    go build

FROM ubuntu:18.04
WORKDIR /app
COPY config.yaml config.yaml
COPY run.sh run.sh
COPY --from=build /app/zilliqa-relayer/zilliqa-relayer zilliqa-relayer
CMD ["/bin/bash"]