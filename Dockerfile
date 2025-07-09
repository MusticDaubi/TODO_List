FROM golang:1.24.1

ARG TODO_PORT
ARG TODO_DBFILE

ENV TODO_PORT=${TODO_PORT}
ENV TODO_DBFILE=${TODO_DBFILE}

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

EXPOSE ${TODO_PORT}

CMD ["/main"]