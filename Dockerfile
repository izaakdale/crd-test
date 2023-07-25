FROM golang:1.20-alpine as builder
WORKDIR /

COPY . .
RUN go mod download

RUN go build -o controller .

FROM scratch
WORKDIR /bin
COPY --from=builder /controller /bin

CMD [ "/bin/controller" ]