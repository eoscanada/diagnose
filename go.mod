module github.com/eoscanada/diagnose

require (
	cloud.google.com/go v0.43.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.6
	github.com/RoaringBitmap/roaring v0.4.16 // indirect
	github.com/abourget/llerrgroup v0.0.0-20161118145731-75f536392d17
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/elastic/gosigar v0.10.4 // indirect
	github.com/eoscanada/bstream v1.6.3-0.20191015134938-44c7bc4effe5
	github.com/eoscanada/derr v0.3.8
	github.com/eoscanada/dhttp v0.0.2-0.20190807044304-212195313a5b
	github.com/eoscanada/dstore v0.0.7
	github.com/eoscanada/dtracing v0.4.2
	github.com/eoscanada/eos-go v0.8.17-0.20191009185653-116355dee341
	github.com/eoscanada/kvdb v0.0.12-0.20191007202513-7f93b1090391
	github.com/eoscanada/logging v0.6.6
	github.com/eoscanada/playground-grpc v0.0.0-20190711150243-551de78c7ae1 // indirect
	github.com/eoscanada/validator v0.4.1-0.20190807042112-8fbbe313c8e8
	github.com/etcd-io/bbolt v1.3.2 // indirect
	github.com/ethereum/go-ethereum v1.9.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/termie/go-shutil v0.0.0-20140729215957-bcacb06fecae // indirect
	github.com/thedevsaddam/govalidator v1.9.6
	github.com/tidwall/gjson v1.3.2
	go.opencensus.io v0.22.1
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20190222213804-5cb15d344471 // indirect
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/klog v0.2.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

go 1.13
