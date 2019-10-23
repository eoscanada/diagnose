module github.com/eoscanada/diagnose

require (
	cloud.google.com/go v0.43.0
	github.com/abourget/llerrgroup v0.0.0-20161118145731-75f536392d17
	github.com/eoscanada/bstream v1.6.3-0.20191022175107-e8fa0d989204
	github.com/eoscanada/derr v0.3.8
	github.com/eoscanada/dhttp v0.0.2-0.20190807044304-212195313a5b
	github.com/eoscanada/dstore v0.0.7
	github.com/eoscanada/dtracing v0.4.2
	github.com/eoscanada/eos-go v0.8.17-0.20191009185653-116355dee341
	github.com/eoscanada/kvdb v0.0.12-0.20191022185346-6acf521538e1
	github.com/eoscanada/logging v0.6.6
	github.com/eoscanada/validator v0.4.1-0.20190807042112-8fbbe313c8e8
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/thedevsaddam/govalidator v1.9.6
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20190222213804-5cb15d344471 // indirect
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/klog v0.2.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

go 1.13
