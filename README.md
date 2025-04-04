# BSV Script Templates

BSV BLOCKCHAIN | Script Templates

A collection of script templates for use with the official BSV Golang SDK

## Overview

The goal of this repository is to provide a place where developers from around the ecosystem can publish all manner of script templates, without needing to update the core library. We're generally neutral and unbiased about what people contribute, so feel free to contribute and see what people do with your cool idea!

## Available Templates

| Template | Description | 
|----------|-------------|
| [BitCom](./template/bitcom/) | BitCom protocol utilities (B, MAP, AIP) for structured data |
| [BSocial](./template/bsocial/) | Social media actions using BitcoinSchema.org standards |
| [BSV20](./template/bsv20/) | BSV20 token standard implementation |
| [BSV21](./template/bsv21/) | BSV21 token standard implementation including LTM and POW20 |
| [Cosign](./template/cosign/) | Co-signing transactions with multiple parties |
| [Inscription](./template/inscription/) | On-chain NFT-like inscriptions |
| [Lockup](./template/lockup/) | Time-locked transactions |
| [OrdLock](./template/ordlock/) | Locking and unlocking functionality for ordinals |
| [OrdP2PKH](./template/ordp2pkh/) | Ordinal-aware P2PKH transactions |
| [P2PKH](./template/p2pkh/) | Standard Pay-to-Public-Key-Hash transactions |
| [Shrug](./template/shrug/) | Experimental template for demo purposes |

Each template folder contains its own README with detailed usage examples.


## Installation

```bash
go get github.com/bitcoin-sv/go-templates
```

## Basic Usage

Import the specific template you need:

```go
import "github.com/bitcoin-sv/go-templates/template/bsocial"
```

See each template's README for detailed examples.

## Contributing

We welcome contributions of all kinds:

1. **Fork & Clone**: Fork this repository and clone it locally
2. **Make Changes**: Add or improve templates
3. **Test**: Ensure all tests pass with `go test`
4. **Document**: Add clear documentation with examples in the template's README
5. **Pull Request**: Submit your changes for review

## License

Open BSV License - See [LICENSE.txt](./LICENSE.txt)

## Support

For questions or issues, please open a GitHub issue or contact the project maintainers.
