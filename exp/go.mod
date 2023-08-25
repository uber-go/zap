module go.uber.org/zap/exp

go 1.19

require (
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.24.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace to commit: https://github.com/uber-go/zap/commit/98e9c4fe632cc00c99033d8d616f1318b7063eee
replace go.uber.org/zap => go.uber.org/zap v1.24.1-0.20230825131617-98e9c4fe632c
