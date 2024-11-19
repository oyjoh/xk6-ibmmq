FROM golang:1.23.2 as builder

# Install necessary packages
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    wget \
    tar

# Install IBM MQ C client libraries
RUN wget https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.4.1.0-IBM-MQC-Redist-LinuxX64.tar.gz \
    && mkdir /opt/mqm \
    && mv 9.4.1.0-IBM-MQC-Redist-LinuxX64.tar.gz /opt/mqm \
    && cd /opt/mqm \
    && tar -xzf 9.4.1.0-IBM-MQC-Redist-LinuxX64.tar.gz \
    && ls -l

# Set environment variables for IBM MQ
ENV LD_LIBRARY_PATH=/opt/mqm/lib64:/opt/mqm/lib
ENV CGO_CFLAGS="-I/opt/mqm/inc"
ENV CGO_LDFLAGS="-L/opt/mqm/lib64 -Wl,-rpath=/opt/mqm/lib64"

WORKDIR /workspace
COPY . .

# Build the k6 binary with the xk6-ibmmq extension
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install go.k6.io/xk6/cmd/xk6@latest
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 xk6 build --with github.com/oyjoh/xk6-ibmmq --output /k6

FROM grafana/k6:latest
COPY --from=builder /opt/mqm/lib64 /opt/mqm/lib64
COPY --from=builder /k6 /usr/bin/k6