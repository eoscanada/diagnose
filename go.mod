module github.com/eoscanada/diagnose

require (
	cloud.google.com/go v0.43.0
	github.com/eoscanada/bstream v1.6.3-0.20191128232437-4b607131f34e
	github.com/eoscanada/derr v0.3.9
	github.com/eoscanada/dgrpc v0.0.0-20191115165705-af05d03bcdcb
	github.com/eoscanada/dhammer v0.0.1
	github.com/eoscanada/dmesh v0.0.0-20191207165858-769f5a73bbff
	github.com/eoscanada/dstore v0.1.4
	github.com/eoscanada/kvdb v0.0.12-0.20191121212719-a7cf2b597a07
	github.com/eoscanada/logging v0.6.6
	github.com/eoscanada/search v0.0.0-20191129050617-aa1cdc9828f2
	github.com/eoscanada/validator v0.4.1-0.20190807042112-8fbbe313c8e8
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/handlers v0.0.0-20181012153334-350d97a79266
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v1.4.1
	github.com/gregjones/httpcache v0.0.0-20190203031600-7a902570cb17 // indirect
	github.com/koding/websocketproxy v0.0.0-20181220232114-7ed82d81a28c
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/thedevsaddam/govalidator v1.9.6
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/zap v1.12.0
	k8s.io/api v0.0.0-20190222213804-5cb15d344471 // indirect
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628 // indirect
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/klog v0.2.0 // indirect
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

go 1.13
