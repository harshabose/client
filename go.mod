module github.com/harshabose/simple_webrtc_comm/client

go 1.24.1

require (
	cloud.google.com/go/firestore v1.18.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/asticode/go-astiav v0.37.0
	github.com/harshabose/mediapipe v0.0.0
	github.com/harshabose/tools v0.0.0
	github.com/pion/interceptor v0.1.40
	github.com/pion/rtp v1.8.19
	github.com/pion/sdp/v3 v3.0.13
	github.com/pion/webrtc/v4 v4.1.2
	google.golang.org/api v0.222.0
	google.golang.org/grpc v1.70.0
)

require (
	cloud.google.com/go v0.117.0 // indirect
	cloud.google.com/go/auth v0.14.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.2.2 // indirect
	cloud.google.com/go/longrunning v0.6.2 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	github.com/asticode/go-astikit v0.52.0 // indirect
	github.com/bluenviron/gortsplib/v4 v4.14.1 // indirect
	github.com/bluenviron/mediacommon/v2 v2.2.0 // indirect
	github.com/coder/websocket v1.8.13 // indirect
	github.com/emirpasic/gods/v2 v2.0.0-alpha // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/pion/datachannel v1.5.10 // indirect
	github.com/pion/dtls/v3 v3.0.6 // indirect
	github.com/pion/ice/v4 v4.0.10 // indirect
	github.com/pion/logging v0.2.3 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.15 // indirect
	github.com/pion/sctp v1.8.39 // indirect
	github.com/pion/srtp/v3 v3.0.5 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/oauth2 v0.26.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250212204824-5a70512c5d8b // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace (
	github.com/harshabose/mediapipe => ../mediapipe
	github.com/harshabose/tools => ../tools
)
