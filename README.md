Diagnose
--------

Small tool to diagnose the in-cluster components, quick internal debug
tools.

Try it locally
-------------

* Set up port-forward on etcd to port 2379
* Start node server
* Start go server

```
kubectl -n dmesh port-forward svc/etcd-client 2379
cd frontend && yarn install && yarn start
go install && diagnose  \
	--api-url=https://mainnet.eos.dfuse.io \
	--blocks-store=gs://dfuseio-global-blocks-us/eos-mainnet/v3 \
	-db-connection=dfuseio-global:dfuse-saas:aca3-v5 \
	-namespace=eos-mainnet \
        -protocol=EOS \
	-search-indexes-store=gs://dfuseio-global-indices-us/eos-mainnet/v2-1/ \
	-dev \
	-skip-k8s
```
