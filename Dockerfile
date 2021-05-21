FROM golang:1.15.2 as build

RUN mkdir /draw-web
WORKDIR /draw-web

COPY . .
RUN if [ ! -d "/draw-web/vendor" ]; then  go mod vendor; fi

RUN make build-in-docker



FROM alpine:3.7

COPY --from=build /draw-web/bin/draw-web /
COPY template /template
CMD ["/draw-web"]