IMAGE:=controller
CLUSTER_NAME:=sample-controller
KIND_VERSION=0.10.0
KUBERNETES_VERSION:=1.20.2
KUBEADM_APIVERSION=kubeadm.k8s.io/v1beta2
ARCH=amd64
OS=$(shell go env GOOS)

.PHONY: setup
setup: api controllers bin tmp

api:
	@mkdir api

controllers:
	@mkdir controllers

bin:
	@mkdir bin

tmp:
	@mkdir tmp

$(IMAGE).img: docker-build
	docker save -o $@ ${IMAGE}

.PHONY: kind
kind:
	if [ ! -f bin/kind ] || [ $$(bin/kind version | awk '{print $$2}') != "v$(KIND_VERSION)" ]; then \
		echo "downloading kind: v$(KIND_VERSION)"; \
		curl -o bin/kind -sfL https://kind.sigs.k8s.io/dl/v$(KIND_VERSION)/kind-$(OS)-$(ARCH) \
		&& chmod a+x bin/kind; \
	fi

.PHONY: launch-kind
launch-kind:
	sed -e "s|@KUBERNETES_VERSION@|$(KUBERNETES_VERSION)|" \
		-e "s|@KUBEADM_APIVERSION@|$(KUBEADM_APIVERSION)|" \
		kind-config.yaml > tmp/kind-config.yaml
	bin/kind create cluster \
		--name=$(CLUSTER_NAME)\
		--config tmp/kind-config.yaml \
		--image kindest/node:v$(KUBERNETES_VERSION)

.PHONY: load-image
load-image:
	rm -f $(IMAGE).img
	$(MAKE) $(IMAGE).img
	bin/kind load image-archive --name=$(CLUSTER_NAME) $(IMAGE).img

.PHONY: shutdown-kind
shutdown-kind:
	bin/kind delete cluster --name=$(CLUSTER_NAME) || true
