FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/ftn-family-guard .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/ftn-family-guard /app/ftn-family-guard
COPY --from=build /src/migrations /app/migrations
COPY --from=build /src/schema.sql /app/schema.sql
EXPOSE 8095
USER nonroot:nonroot
ENTRYPOINT ["/app/ftn-family-guard"]
