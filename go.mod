module github.com/ripta/netdebug

go 1.23.4

toolchain go1.23.8

require (
	github.com/coreos/go-oidc/v3 v3.14.1
	github.com/miekg/dns v1.1.66
	github.com/ripta/rt v0.0.0-20250409051646-3283bd3d0519
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
	github.com/thediveo/enumflag/v2 v2.0.7
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/net v0.40.0
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6
	k8s.io/klog/v2 v2.130.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/oauth2 v0.29.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-jose/go-jose/v4 => github.com/go-jose/go-jose/v4 v4.0.5
