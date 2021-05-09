# k8s-sample-controller

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
