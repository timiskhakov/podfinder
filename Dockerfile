# Stage 1: Build

FROM golang:1.18-alpine as build

RUN apk --no-cache add gcc libc-dev

ADD app /build/app
COPY go.mod /build
COPY go.sum /build

WORKDIR /build

RUN go mod download

RUN cd app && go test ./...
RUN cd app && go build -o podfinder

# Stage 2: App

FROM alpine:3.15

WORKDIR /srv

COPY --from=build /build/app/podfinder /srv/podfinder
COPY --from=build /build/app/templates /srv/templates
COPY --from=build /build/app/www /srv/www

EXPOSE 3000

CMD ["/srv/podfinder"]