# k8s-sample-controller

## デプロイ(kind)

```
$ make setup
$ make launch-kind
$ make load-image
$ make deploy
```

### サンプルのデプロイ

```
$ kubectl apply -f config/samples/hello_v1beta1_message.yaml
```

確認

```
$ kubectl get -f config/samples/hello_v1beta1_message.yaml
NAME             HELLOMESSAGE
message-sample   Hello bar
```

## テスト

```
$ make setup
$ make launch-kind
$ make load-image
$ make deploy
$ make e2e
```

## 作業手順

### 事前準備

- Go言語のv1.15系バージョンを事前にインストールする
  https://golang.org/doc/install

  設定例:
  ```
  $ wget https://golang.org/dl/go1.15.12.linux-amd64.tar.gz
  $ rm -rf /usr/local/go && tar -C /usr/local -xzf go1.15.12.linux-amd64.tar.gz
  $ echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.bash_profile
  ```
- GOPATHの設定
  設定例:
  ```
  $ echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.bash_profile
  ```

- dockerのインストール
  設定例:
  ```
  $ sudo apt-get update && apt-get install docker.io
  $ sudo systemctl enable docker
  ```

## kubebuilderインストール~セットアップ

https://book.kubebuilder.io/quick-start.html

```
$ curl -o /usr/local/bin/kubebuilder -sfL https://go.kubebuilder.io/dl/3.0.0/$(go env GOOS)/$(go env GOARCH) && chmod a+x /usr/local/bin/kubebuilder
$ kubebuilder init --domain bells17.io --repo github.com/bells17/k8s-sample-controller
```

## 動作確認用の便利ファイルの生成

kindでサクッと動作確認をしたいのでそれらを実行するためのMakefileを作ったりしてる

```
$ cat << EOS >> Makefile

include controller.mk
EOS

$ cat << EOS > controller.mk
IMAGE:=controller
CLUSTER_NAME:=sample-controller
KIND_VERSION=0.10.0
KUBERNETES_VERSION:=1.20.2
KUBEADM_APIVERSION=kubeadm.k8s.io/v1beta2
ARCH=amd64
OS=$(shell go env GOOS)

.PHONY: setup
setup: $(shell pwd)/api $(shell pwd)/controllers $(shell pwd)/bin $(shell pwd)/tmp ginkgo kind kubectl

$(shell pwd)/api:
	@mkdir $(shell pwd)/api

$(shell pwd)/controllers:
	@mkdir $(shell pwd)/controllers

$(shell pwd)/bin:
	@mkdir $(shell pwd)/bin

$(shell pwd)/tmp:
	@mkdir $(shell pwd)/tmp

.PHONY: ginkgo
ginkgo:
	$(call go-get-tool,bin/ginkgo,github.com/onsi/ginkgo/ginkgo@v1.14.1)

.PHONY: kind
kind:
	if [ ! -f $(shell pwd)/bin/kind ] || [ $$($(shell pwd)/bin/kind version | awk '{print $$2}') != "v$(KIND_VERSION)" ]; then \
		echo "downloading kind: v$(KIND_VERSION)"; \
		curl -o $(shell pwd)/bin/kind -sfL https://kind.sigs.k8s.io/dl/v$(KIND_VERSION)/kind-$(OS)-$(ARCH) \
		&& chmod a+x $(shell pwd)/bin/kind; \
	fi

.PHONY: kubectl
kubectl:
	@if [ ! -f /usr/local/bin/kubectl ] || [ $$(/usr/local/bin/kubectl version --client=true --short=true | awk '{print $$3}') != "v$(KUBERNETES_VERSION)" ]; then \
		echo "downloading kubectl: v$(KUBERNETES_VERSION)"; \
		curl -o /usr/local/bin/kubectl -sfL https://storage.googleapis.com/kubernetes-release/release/v$(KUBERNETES_VERSION)/bin/$(OS)/$(ARCH)/kubectl && chmod a+x /usr/local/bin/kubectl; \
	fi

$(IMAGE).img: docker-build
	docker save -o $@ ${IMG}

.PHONY: launch-kind
launch-kind: shutdown-kind
	sed -e "s|@KUBERNETES_VERSION@|$(KUBERNETES_VERSION)|" \
		-e "s|@KUBEADM_APIVERSION@|$(KUBEADM_APIVERSION)|" \
		kind-config.yaml > $(shell pwd)/tmp/kind-config.yaml
	$(shell pwd)/bin/kind create cluster \
		--name=$(CLUSTER_NAME)\
		--config $(shell pwd)/tmp/kind-config.yaml \
		--image kindest/node:v$(KUBERNETES_VERSION)

.PHONY: load-image
load-image:
	rm -f $(IMAGE).img
	$(MAKE) $(IMAGE).img
	$(shell pwd)/bin/kind load image-archive --name=$(CLUSTER_NAME) $(IMAGE).img

.PHONY: shutdown-kind
shutdown-kind:
	$(shell pwd)/bin/kind delete cluster --name=$(CLUSTER_NAME) || true

.PHONY: e2e
e2e:
	kubectl -n k8s-sample-controller-system wait deploy k8s-sample-controller-controller-manager --for condition=Progressing
	kubectl create ns test || true
	E2E=true $(shell pwd)/bin/ginkgo -timeout=1h --failFast -v controllers/
	kubectl delete ns test || true

.PHONY: clean
clean: shutdown-kind
	rm -fr $(IMAGE).img $(shell pwd)/bin $(shell pwd)/tmp

EOS

$ cat << EOS > kind-config.yaml
apiVersion:  kind.x-k8s.io/v1alpha4
kind: Cluster
kubeadmConfigPatches:
- |
  apiVersion: "@KUBEADM_APIVERSION@"
  kind: ClusterConfiguration
  metadata:
    name: config
  kubernetesVersion: "v@KUBERNETES_VERSION@"
EOS
```

## 開発フロー

- `make launch-kind` でkindを起動
- 開発する
- `kind load-image` でビルドしたイメージをkindに読み込み
- `make deploy` でマニフェストをデプロイする

π