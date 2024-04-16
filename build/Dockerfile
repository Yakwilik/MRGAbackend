FROM golang:alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /src/cmd/app ./cmd/...

FROM alpine:latest
COPY --from=build /src/cmd/app /cmd/app

ENTRYPOINT ["/cmd/app"]