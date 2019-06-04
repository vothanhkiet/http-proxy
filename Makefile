REGISTRY_URL ?= vothanhkiet/http-proxy
TAG ?= 1.0.1

build:
	docker build -f docker/Dockerfile --squash -t ${REGISTRY_URL}:latest -t ${REGISTRY_URL}:${TAG} .

upload: 
	docker push ${REGISTRY_URL}:latest
	docker push ${REGISTRY_URL}:${TAG}