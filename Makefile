TARGET=DRAGON
GO=go
GO_MODULE=GO111MODULE=on
BIN_DIR=bin/
ALPINE_COMPILE_FLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
PACKAGE_PREFIX=github.com/NTHU-LSALAB/DRAGON/cmd/

.PHONY: all clean $(TARGET)

all: $(TARGET)

DRAGON:
	$(GO_MODULE) $(ALPINE_COMPILE_FLAGS) $(GO) build -o $(BIN_DIR)$@ $(PACKAGE_PREFIX)$@

clean:
	rm -r bin 2>/dev/null; exit 0

install:
	kubectl create -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml
	kubectl create -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/dragon.yaml

uninstall:
	kubectl delete -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/dragon.yaml
	kubectl delete -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml

install-custom:
	kubectl create -f ./config/monitor.yaml
	kubectl create -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml
	kubectl create -f ./config/dragon.yaml

uninstall-custom:
	kubectl delete -f ./config/dragon.yaml
	kubectl delete -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml
	kubectl delete -f ./config/monitor.yaml

release:
	docker build -t haverzard/dragon:0.0.0 -f docker/DRAGON/Dockerfile .
	docker push haverzard/dragon:0.0.0

release-api:
	docker build -t haverzard/monitor-api:0.0.0 -f experiments/monitor-api/Dockerfile experiments/monitor-api/
	docker push haverzard/monitor-api:0.0.0

release-api-v2:
	docker build -t haverzard/monitor-api:go-0.0.1 -f experiments/monitor-api-v2/Dockerfile experiments/monitor-api-v2/
	docker push haverzard/monitor-api:go-0.0.1

release-tf-image:
	docker build -t haverzard/tf-image:0.0.0 -f experiments/jobs/Dockerfile experiments/jobs/
	docker push haverzard/tf-image:0.0.0

init-cluster:
	minikube start --nodes 3 -p ta-playground
	minikube addons enable metrics-server -p ta-playground

delete-cluster:
	minikube stop -p ta-playground

test:
	kubectl apply -f experiments/jobs/job1-v2.yaml
	kubectl apply -f experiments/jobs/job2-v2.yaml
	kubectl apply -f experiments/jobs/job3-v2.yaml

reset:
	kubectl delete -f experiments/jobs/job1-v2.yaml
	kubectl delete -f experiments/jobs/job2-v2.yaml
	kubectl delete -f experiments/jobs/job3-v2.yaml