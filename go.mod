module github.com/bitcoin-sv/go-templates

go 1.24.1

require github.com/bsv-blockchain/go-sdk v0.0.0-00010101000000-000000000000

require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.35.0 // indirect
)

replace github.com/bsv-blockchain/go-sdk => ../go-sdk
