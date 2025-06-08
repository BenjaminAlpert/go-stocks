dev-build:
	go build -o ./bin/go-stocks

dev-run:
	TIINGO_TOKEN="op://k8s/Tiingo/token" op run -- ./bin/go-stocks

dev-build-and-run: dev-build dev-run

REGISTRY_HOST=registry.a0-0.com
REGISTRY_IMAGE_PATH=/go-stocks
REGISTRY_IMAGE_TAG=latest

REGISTRY_HELM_CHART_PATH=/charts/go-stocks
REGISTRY_HELM_CHART_TAG:=$(shell yq '.version' ./helm/chart/Chart.yaml)

image-build:
	docker build -t ${REGISTRY_HOST}${REGISTRY_IMAGE_PATH}:${REGISTRY_IMAGE_TAG} . --platform=linux/amd64

image-push:
	docker push ${REGISTRY_HOST}${REGISTRY_IMAGE_PATH}:${REGISTRY_IMAGE_TAG}

helm-push:
	helm package ./helm/chart
	helm push "go-stocks-${REGISTRY_HELM_CHART_TAG}.tgz" oci://${REGISTRY_HOST}${REGISTRY_HELM_CHART_PATH}
	rm "go-stocks-${REGISTRY_HELM_CHART_TAG}.tgz"

redeploy:
	kubectl delete pods -n go-stocks -l app=go-stocks
	kubectl wait pods -n go-stocks -l app=go-stocks --for='jsonpath={.status.conditions[?(@.type=="Ready")].status}=True'
	kubectl logs -n go-stocks -l app=go-stocks

build: image-build

push: image-push helm-push

build-and-push: build push

build-and-push-and-redeploy: build-and-push redeploy