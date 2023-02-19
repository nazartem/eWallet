FROM golang:1.19-buster as builder
WORKDIR /app/

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o ./e-wallet ./cmd/.

FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /app/e-wallet ./

EXPOSE 8000

RUN mkdir './data' './data/sqlite' './logs'
#RUN mkdir './logs
RUN touch './logs/all.logs'

CMD ["./e-wallet"]