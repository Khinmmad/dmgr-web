# Build
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /dmgr-web-backend .

# Run
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /dmgr-web-backend /app/dmgr-web-backend
ENV PORT=8080 DATA_DIR=/data
RUN mkdir -p /data && chown app /data
USER app
VOLUME ["/data"]
EXPOSE 8080
CMD ["/app/dmgr-web-backend"]
