module github.com/Terry-Mao/goim

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Shopify/sarama v1.19.0 // indirect
	github.com/bilibili/discovery v1.0.1
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/gogo/protobuf v1.1.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.2.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.0.0
	github.com/issue9/assert v1.0.0
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.8.0
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a
	github.com/stretchr/testify v1.3.0
	github.com/thinkboy/log4go v0.0.0-20160303045050-f91a411e4a18
	github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1
	google.golang.org/grpc v1.16.0
	gopkg.in/Shopify/sarama.v1 v1.19.0
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190227174305-5b3e6a55c961
	golang.org/x/net => github.com/golang/net v0.0.0-20181011144130-49bb7cea24b1
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180830151530-49385e6e1522
	golang.org/x/text => github.com/golang/text v0.3.0
	golang.org/x/tools => github.com/golang/tools v0.0.0-20180828015842-6cd1fcedba52
	google.golang.org/appengine => github.com/golang/appengine v1.1.0
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20180817151627-c66870c02cf8
	google.golang.org/grpc => github.com/grpc/grpc-go v1.16.0
)
