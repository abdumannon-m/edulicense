FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/edu-license ./cmd/server

FROM alpine:3.22

RUN adduser -D -h /app appuser
WORKDIR /app

COPY --from=build /out/edu-license /app/edu-license
COPY internal/templates /app/internal/templates
COPY web/static /app/web/static
COPY migrations /app/migrations

USER appuser
EXPOSE 8080

CMD ["/app/edu-license", "serve"]
