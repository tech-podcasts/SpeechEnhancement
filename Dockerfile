FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY . /app


RUN go build -o main

FROM python:3.11-alpine


RUN wget https://github.com/Rikorose/DeepFilterNet/releases/download/v0.4.0/deep-filter-v0.4.0-x86_64-unknown-linux-musl -O /usr/local/bin/deep-filter && chmod +x /usr/local/bin/deep-filter

RUN apk update

RUN apk add --no-cache ffmpeg

RUN pip install --no-cache-dir ffmpeg-normalize

WORKDIR /app

RUN mkdir -p /app/uploads

ENV GIN_MODE release

COPY --from=builder /app/main /app/main
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/dist /app/dist

EXPOSE 8080

ENTRYPOINT ["/app/main"]

