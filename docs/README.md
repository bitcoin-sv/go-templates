# BSV Script Templates Documentation

Welcome to the documentation for the BSV Script Templates repository. This documentation provides detailed information about the various Bitcoin SV script templates available in this repository.

## Structure

- Templates `./templates/*/README.md` - Documentation for individual script templates
- [Contributing](../CONTRIBUTING.md) - Guide for contributors

## Getting Started

### Installation

```bash
go get github.com/bitcoin-sv/go-templates
```

### Basic Usage

1. Import the specific template you need:
   ```go
   import "github.com/bitcoin-sv/go-templates/template/bsocial"
   ```

2. Use the template functions to create or decode transactions:
   ```go
   // See individual template documentation for specific usage examples
   ```

## Available Templates

The repository includes templates for various use cases:

- **[BitCom](./templates/bitcom/README.md)** - BitCom protocol utilities (B, MAP, AIP)
- **[BSocial](./templates/bsocial/README.md)** - Social media actions using BitcoinSchema.org standards
- **[BSV20](./templates/bsv20/README.md)** - BSV20 token standard implementation
- **[BSV21](./templates/bsv21/README.md)** - BSV21 token standard implementation including LTM and POW20
- **[Cosign](./templates/cosign/README.md)** - Co-signing transactions with multiple parties
- **[Inscription](./templates/inscription/README.md)** - On-chain NFT-like inscriptions
- **[Lockup](./templates/lockup/README.md)** - Time-locked transactions
- **[OrdLock](./templates/ordlock/README.md)** - Locking and unlocking functionality for ordinals
- **[OrdP2PKH](./templates/ordp2pkh/README.md)** - Ordinal-aware P2PKH transactions
- **[P2PKH](./templates/p2pkh/README.md)** - Standard Pay-to-Public-Key-Hash transactions
- **[Shrug](./templates/shrug/README.md)** - Experimental template for demo purposes

## Support

If you encounter issues or have questions, please:

1. Check the documentation for the specific template
2. Search existing GitHub issues
3. Open a new issue if needed

## License

The code in this repository is licensed under the Open BSV License. See [LICENSE.txt](../LICENSE.txt) for details. 