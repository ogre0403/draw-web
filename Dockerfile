FROM golang:1.15.7-alpine3.13 as build

RUN mkdir /draw-web
WORKDIR /draw-web

COPY . .
RUN if [ ! -d "/draw-web/vendor" ]; then  go mod vendor; fi

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o bin/draw-web cmd/*.go



FROM alpine:3.13
RUN apk add --no-cache tzdata ca-certificates && update-ca-certificates
COPY --from=build /draw-web/bin/draw-web /
COPY template /template
ENTRYPOINT ["/draw-web"]