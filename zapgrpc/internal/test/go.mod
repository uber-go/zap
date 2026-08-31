module go.uber.org/zap/zapgrpc/internal/test

go 1.25.0

require (
	github.com/stretchr/testify v1.12.1
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.83.2
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace go.uber.org/zap => ../../..
