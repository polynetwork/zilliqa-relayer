FROM golang:1.15 AS build
WORKDIR /app/zilliqa-relayer
COPY . ./
RUN go build
#RUN git clone https://github.com/Zilliqa/zilliqa-relayer.git  && \
#    cd zilliqa-relayer && \
#    go build

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
COPY run.sh run.sh
COPY --from=build /app/zilliqa-relayer/zilliqa-relayer zilliqa-relayer
ENTRYPOINT ["/bin/bash", "run.sh"]
