FROM golang:1.24 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make

FROM gcr.io/distroless/static-debian11
COPY --chmod=555 --from=builder /usr/src/app/bin/linux/app /
VOLUME /cfg
EXPOSE 80
ENTRYPOINT ["/app", "-addr", ":80", "-addressbooksfile", "/cfg/config.json"]
