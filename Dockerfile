# syntax=docker/dockerfile:1
FROM golang:1.20 AS build
WORKDIR /go/src/github.com/itsdalmo/tfcheck
ADD . /go/src/github.com/itsdalmo/tfcheck
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o /go/bin/tfcheck -trimpath -ldflags="-buildid=\"\" -s -w" -v cmd/tfcheck/main.go

FROM hashicorp/terraform:1.4.5
COPY --from=build /go/bin/tfcheck /
ENTRYPOINT ["/tfcheck"]
