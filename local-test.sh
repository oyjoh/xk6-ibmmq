#!/usr/bin/env bash
set -euo pipefail

# Build
export CGO_ENABLED=1
go install go.k6.io/xk6/cmd/xk6@latest
xk6 build \
    --with github.com/iambaim/xk6-ibmmq=.

# Run dev MQ container and wait until MQ is ready
docker-compose -f example/docker-compose.yml up -d localmqtest
sleep 10

# Run tests
export MQ_QMGR="QM1"
export MQ_CHANNEL="DEV.APP.SVRCONN"
export MQ_HOST="localhost"
export MQ_PORT=1414
export MQ_USERID="app"
export MQ_PASSWORD="password"

./k6 run --vus 2 --duration 5s example/localtest.js

docker-compose -f example/docker-compose.yml down