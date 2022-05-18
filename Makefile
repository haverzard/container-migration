TARGET=DRAGON
GO=go
GO_MODULE=GO111MODULE=on
BIN_DIR=bin/
ALPINE_COMPILE_FLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
PACKAGE_PREFIX=github.com/NTHU-LSALAB/DRAGON/cmd/
TF_IMAGE_VERSION=0.0.0

.PHONY: all clean $(TARGET)

all: $(TARGET)

DRAGON:
	$(GO_MODULE) $(ALPINE_COMPILE_FLAGS) $(GO) build -o $(BIN_DIR)$@ $(PACKAGE_PREFIX)$@

clean:
	rm -r bin 2>/dev/null; exit 0

install:
	kubectl create -f ./deployments/kubernetes/crd.yaml
	kubectl create -f ./deployments/kubernetes/dragon-orig.yaml

uninstall:
	kubectl delete -f ./deployments/kubernetes/dragon-orig.yaml
	kubectl delete -f ./deployments/kubernetes/crd.yaml

install-custom:
	kubectl create -f ./deployments/kubernetes/monitor.yaml
	kubectl create -f ./deployments/kubernetes/crd.yaml
	kubectl create -f ./deployments/kubernetes/dragon.yaml

uninstall-custom:
	kubectl delete -f ./deployments/kubernetes/dragon.yaml
	kubectl delete -f ./deployments/kubernetes/crd.yaml
	kubectl delete -f ./deployments/kubernetes/monitor.yaml

release-dragon:
	docker build -t haverzard/dragon:$(VERSION) -f ./deployments/docker/DRAGON/Dockerfile .
	docker push haverzard/dragon:$(VERSION)

release-monitor:
	docker build -t haverzard/container-monitor:$(VERSION) -f ./deployments/docker/container-monitor/Dockerfile ./internal/container-monitor/
	docker push haverzard/container-monitor:$(VERSION)

release-tf-image:
	docker build -t haverzard/tf-image:$(TF_IMAGE_VERSION) -f ./deployments/docker/tf-image/Dockerfile .
	docker push haverzard/tf-image:$(TF_IMAGE_VERSION)

init-local-cluster:
	minikube start --nodes 3 -p ta-playground
	minikube start --nodes 1 -p ta-playground --kubernetes-version=v.1.19.16
	minikube addons enable metrics-server -p ta-playground

delete-local-cluster:
	minikube stop -p ta-playground

URL ?= 'https://raw.githubusercontent.com/haverzard/container-migration/main/internal/jobs/speed/mnist-df.py'
MAX_REPLICAS ?= 1
MIN_REPLICAS ?= 1
INIT_REPLICAS ?= 1
OUTPUT_FILE ?= 'deployments/kubernetes/jobs/example.yaml'
gen-tc:
	TF_IMAGE_VERSION=$(TF_IMAGE_VERSION) ./scripts/generate-tc.sh $(URL) $(MAX_REPLICAS) $(MIN_REPLICAS) $(INIT_REPLICAS) $(OUTPUT_FILE)

test:
	./scripts/substitute-test.sh ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job1.yaml | kubectl apply -f -
	./scripts/substitute-test.sh ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job2.yaml | kubectl apply -f -
	./scripts/substitute-test.sh ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job3.yaml | kubectl apply -f -

reset:
	kubectl delete -f ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job1.yaml
	kubectl delete -f ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job2.yaml
	kubectl delete -f ./deployments/kubernetes/jobs/$(SYSTEM)/$(SCENARIO)/job3.yaml
