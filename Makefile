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
	kubectl create -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml
	kubectl create -f ./config/dragon.yaml

uninstall-custom:
	kubectl delete -f ./config/dragon.yaml
	kubectl delete -f https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON/v0.9/crd.yaml

release:
	docker build -t haverzard/dragon:0.0.0 -f docker/DRAGON/Dockerfile .
	docker push haverzard/dragon:0.0.0

init-cluster:
	kind create cluster --name k8s-playground --config config/kind-config.yaml