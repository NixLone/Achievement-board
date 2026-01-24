FROM golang:1.22-alpine AS build
WORKDIR /app

# Render sometimes sets -mod=readonly; allow module downloads to proceed.
ENV GOFLAGS=-mod=mod

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/server /app/server
# include migrations so server can bootstrap schema on startup
COPY --from=build /app/migrations /app/migrations

EXPOSE 8080
CMD ["/app/server"]
