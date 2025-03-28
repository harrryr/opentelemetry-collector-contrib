module github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/xray

go 1.22.4

require (
	github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws v0.0.0-20241107200025-113eee2da3ff
	github.com/aws/aws-sdk-go v1.53.11
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil v0.103.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector/component v0.103.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.103.0 // indirect
	go.opentelemetry.io/collector/pdata v1.10.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil => ../../../internal/aws/awsutil

replace github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws => ../../../override/aws

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
