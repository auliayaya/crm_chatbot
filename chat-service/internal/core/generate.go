package core

//go:generate mockery --dir ./internal/core/ports --name ".*Repository" --output ./internal/mocks --outpkg mocks
//go:generate mockery --dir ./internal/core/ports --name ".*Service" --output ./internal/mocks --outpkg mocks
//go:generate mockery --dir ./internal/core/ports --name ".*Publisher" --output ./internal/mocks --outpkg mocks