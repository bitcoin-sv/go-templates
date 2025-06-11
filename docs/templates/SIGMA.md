# Sigma Protocol Template

This package implements a Go version of the Sigma digital signature protocol for Bitcoin SV. Sigma is a signature scheme for signing Bitcoin transaction data, which can be used to prove ownership or authenticity of data stored on the blockchain.

## Overview

The Sigma protocol allows for:
- Creating digital signatures with various algorithms
- Multiple signature formats (string, binary, hex)
- Optional message and nonce fields
- Verification of signatures

## Usage

### Decoding Sigma Signatures from Scripts

```go
import (
    "github.com/bitcoin-sv/go-templates/template/bitcom"
    "github.com/bitcoin-sv/go-templates/template/sigma"
)

// Decode the transaction script
b := bitcom.DecodeTx(tx)

// Extract Sigma signatures
signatures := sigma.Decode(b)

// Process signatures
for _, sig := range signatures {
    // Get the signature bytes based on the signature type
    sigBytes, err := sig.GetSignatureBytes()
    if err != nil {
        // Handle error
    }
    
    // Use the signature bytes for verification or other purposes
    // ...
}
```

## Protocol Details

The Sigma protocol uses the following format in scripts:

```
OP_RETURN
1signaturezzYYzQ2H2st5SvdT9KwGe  # Sigma protocol prefix
<algorithm>                       # e.g., "ECDSA", "SHA256-ECDSA"
<signer_address>                  # Bitcoin address of the signer
<signature_value>                 # The actual signature
[<signature_type>]                # Optional: "string", "binary", or "hex"
[<message>]                       # Optional: the message that was signed
[<nonce>]                         # Optional: random nonce used for signing
```

### Required Fields

- **Algorithm**: The signature algorithm used (e.g., "ECDSA", "SHA256-ECDSA")
- **Signer Address**: The Bitcoin address of the signer
- **Signature Value**: The signature itself

### Optional Fields

- **Signature Type**: The format of the signature ("string", "binary", "hex")
  - Defaults to "string" if not specified
- **Message**: The content that was signed
- **Nonce**: A random value used during signature generation

## Examples

See the test file for examples of creating and decoding Sigma signatures.

## Reference

This implementation is based on the [Sigma protocol specification](https://github.com/BitcoinSchema/go-sigma) and is compatible with the JavaScript Sigma Library. 