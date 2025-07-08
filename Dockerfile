FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

EXPOSE 7540

ENV TODO_DBFILE="scheduler.db"
ENV TODO_PORT=7540

CMD ["/main"]