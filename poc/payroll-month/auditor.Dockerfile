# Auditor image (UNEXECUTED: needs a docker daemon). Build from repo root:
#   docker build -f poc/payroll-month/auditor.Dockerfile .
FROM golang:1.27-bookworm AS build
WORKDIR /src
# The POC module replaces github.com/PithomLabs/a2 with ../.., so module
# resolution needs the root go.mod even though the auditor imports no
# main-module packages.
COPY go.mod go.sum ./
COPY poc/payroll-month/go.mod poc/payroll-month/go.sum ./poc/payroll-month/
RUN cd poc/payroll-month && go mod download
COPY poc/payroll-month/ ./poc/payroll-month/
RUN cd poc/payroll-month && CGO_ENABLED=0 go build -o /auditor ./cmd/auditor

FROM debian:bookworm-slim
RUN useradd -m auditor
USER auditor
COPY --from=build /auditor /auditor
ENTRYPOINT ["/auditor"]
