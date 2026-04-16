FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/agent .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/agent /app/agent
USER nonroot:nonroot
ENTRYPOINT ["/app/agent"]
