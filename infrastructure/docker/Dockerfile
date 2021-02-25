FROM golang:1.16 AS build

RUN apt update && apt install -y ffmpeg sqlite3 libsqlite3-dev
WORKDIR /go/src/github.com/jan-carreras/torrentpreviewer
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY http ./
RUN CGO_ENABLED=1 go build --tags "libsqlite3 linux" -o /go/bin/http ./cmd/http/http.go
RUN CGO_ENABLED=1 go build --tags "libsqlite3 linux" -o /go/bin/torrentprev ./cmd/cli/torrentprev/torrentprev.go


FROM build as http
ENTRYPOINT ["/go/bin/http"]


FROM build as cli
