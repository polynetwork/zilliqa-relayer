FROM golang:1.15 AS build
WORKDIR /app
RUN git clone https://github.com/Zilliqa/zilliqa-relayer.git  && \
    cd zilliqa-relayer && \
    go build

FROM ubuntu:18.04
WORKDIR /app
COPY config.yaml config.yaml
COPY --from=build /app/zilliqa-relayer run
CMD ["/bin/bash"]