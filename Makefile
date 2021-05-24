PROJ_NAME = draw-web
REGISTRY = ogre0403
TAG = latest
IMAGE = $(REGISTRY)/$(PROJ_NAME):$(TAG)

build:
	rm -rf bin/${PROJ_NAME}
	go build -mod=vendor  \
	-o bin/${PROJ_NAME} cmd/*.go

run:
	bin/${PROJ_NAME} -alsologtostderr -password=${PASSWORD} -mail=${MAIL}

build-img:
	docker build -t $(IMAGE) -f ./Dockerfile .


run-in-docker:
	docker run --rm -ti -p 8080:8080 $(REGISTRY)/$(PROJ_NAME):$(TAG)

clean:
	rm -rf bin/
	rm -f *.csv
