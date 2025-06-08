build-dev:
	go build -o ./bin/go-stocks

run-dev:
	TIINGO_TOKEN="op://k8s/Tiingo/token" op run -- ./bin/go-stocks

build-and-run-dev: build-dev run-dev

REGISTRY_HOST=registry.a0-0.com
REGISTRY_PATH=/go-stocks
REGISTRY_TAG=latest


build:
	docker build -t ${REGISTRY_HOST}${REGISTRY_PATH}:${REGISTRY_TAG} . --platform=linux/amd64

push:
	docker push ${REGISTRY_HOST}${REGISTRY_PATH}:${REGISTRY_TAG}

build-and-push: build push

build-and-push-and-redeploy: build-and-push
	kubectl delete pods -n go-stocks -l app=go-stocks
	kubectl wait pods -n go-stocks -l app=go-stocks --for='jsonpath={.status.conditions[?(@.type=="Ready")].status}=True'
	kubectl logs -n go-stocks -l app=go-stocks