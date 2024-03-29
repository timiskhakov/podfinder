# Stage 1: Build

FROM golang:1.20-alpine as build

RUN apk --no-cache add gcc libc-dev

ENV GOFLAGS="-mod=vendor"

ADD app /build/app
ADD vendor /build/vendor
COPY go.mod /build
COPY go.sum /build

WORKDIR /build

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.52.2
RUN ./bin/golangci-lint run
RUN cd app && go test ./...
RUN cd app && go build -o podfinder

# Stage 2: Run

FROM alpine:3.15

WORKDIR /srv

COPY --from=build /build/app/podfinder /srv/podfinder
COPY --from=build /build/app/templates /srv/templates
COPY --from=build /build/app/www /srv/www

EXPOSE 3000

CMD ["/srv/podfinder"]