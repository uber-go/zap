module go.uber.org/zap/exp

go 1.20

require (
	github.com/stretchr/testify v1.12.1
	go.uber.org/zap v1.26.0
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace go.uber.org/zap => ../
