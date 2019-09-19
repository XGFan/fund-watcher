FROM golang:alpine AS builder
WORKDIR /app
ADD . ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"  -o fund .
RUN ls -al

FROM alpine:3
COPY --from=builder /app/fund /app/fund
EXPOSE 16000
ENTRYPOINT ["/app/fund"]