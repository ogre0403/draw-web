PROJ_NAME = draw-web
REGISTRY = ogre0403
TAG = latest
IMAGE = $(REGISTRY)/$(PROJ_NAME):$(TAG)

build:
	rm -rf bin/${PROJ_NAME}
	go build -mod=vendor  \
	-o bin/${PROJ_NAME} cmd/main.go

run:
	bin/${PROJ_NAME} -alsologtostderr

build-img:
	docker build -t $(IMAGE) -f ./Dockerfile .

build-in-docker:
	rm -rf bin/${PROJ_NAME}
	CGO_ENABLED=0 GOOS=linux \
	go build -mod=vendor \
	-o bin/${PROJ_NAME} cmd/main.go

run-in-docker:
	docker run --rm -ti -p 8080:8080 $(REGISTRY)/$(PROJ_NAME):$(TAG)

clean:
	rm *.csv
	rm -rf bin/