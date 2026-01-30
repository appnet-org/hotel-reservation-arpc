module github.com/appnetorg/hotel-reservation-arpc

go 1.24.0

require (
	github.com/appnet-org/arpc v0.0.0-20260127065040-422057d11fe2
	github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto v0.0.0-00010101000000-000000000000
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/google/uuid v1.6.0
	github.com/hailocab/go-geoindex v0.0.0-20160127134810-64631bfe9711
	github.com/hashicorp/consul v1.0.6
	github.com/opentracing-contrib/go-stdlib v0.0.0-20180308002341-f6b9967a3c69
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rs/zerolog v1.31.0
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	google.golang.org/grpc v1.75.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.1 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/hashicorp/go-cleanhttp v0.0.0-20171218145408-d5fe4b57a186 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-rootcerts v0.0.0-20160503143440-6bb64b370b90 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/hashicorp/serf v0.8.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/go-homedir v0.0.0-20161203194507-b8bc1bf76747 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/appnetorg/hotel-reservation-arpc/services/hotel/proto => ./proto/hotel
