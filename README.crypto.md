# Jot Cryptography Implementation

## Overview

Jot uses modern cryptographic primitives from the Go standard library and `golang.org/x/crypto` packages to secure journal entries. This document details the cryptographic implementation, key management, and security considerations.

## Key Components

### Libraries Used
- `golang.org/x/crypto/nacl/box`: For public-key cryptography (asymmetric encryption)
- `golang.org/x/crypto/nacl/secretbox`: For symmetric encryption
- `crypto/rand`: For secure random number generation

### Key Types and Storage

#### Master Key Pair
- A NaCl key pair (public/private) is generated for each journal
- Private key: 32-byte random value
- Public key: 32-byte derived from private key
- Stored in `$HOME/.jot/backup/`:
  - `jot.pub`: Public key in Base64 format
  - `jot.sec`: Private key in Base64 format (permissions: 0600)

### Encryption Process

1. **Journal Creation**
   ```go
   // Generate new key pair
   publicKey, privateKey, err := box.GenerateKey(rand.Reader)
   ```

2. **Entry Encryption**
   ```go
   // For each entry:
   // 1. Generate random 24-byte nonce
   // 2. Encrypt entry text using box.Seal
   // 3. Store nonce + encrypted data
   ```

3. **Entry Decryption**
   ```go
   // 1. Extract nonce from stored data
   // 2. Decrypt using box.Open
   ```

## Implementation Details

### Key Generation
```go
func GenerateKey() (string, error) {
    // Generate NaCl key pair
    publicKey, privateKey, err := box.GenerateKey(rand.Reader)
    if err != nil {
        return "", fmt.Errorf("failed to generate key pair: %w", err)
    }
    
    // Convert to storable format
    pubKeyStr := base64.StdEncoding.EncodeToString(publicKey[:])
    privKeyStr := base64.StdEncoding.EncodeToString(privateKey[:])
    
    // Store keys
    if err := backupKey(pubKeyStr, privKeyStr); err != nil {
        return "", err
    }
    
    return pubKeyStr, nil
}
```

### Entry Encryption
```go
func Encrypt(text string, publicKey []byte) ([]byte, error) {
    // Generate random nonce
    var nonce [24]byte
    if _, err := rand.Read(nonce[:]); err != nil {
        return nil, fmt.Errorf("nonce generation failed: %w", err)
    }
    
    // Encrypt message
    encrypted := box.Seal(nonce[:], []byte(text), &nonce, publicKey, privateKey)
    return encrypted, nil
}
```

### Entry Decryption
```go
func Decrypt(data []byte, publicKey, privateKey []byte) (string, error) {
    // Extract nonce
    var nonce [24]byte
    copy(nonce[:], data[:24])
    
    // Decrypt message
    decrypted, ok := box.Open(nil, data[24:], &nonce, publicKey, privateKey)
    if !ok {
        return "", fmt.Errorf("decryption failed")
    }
    
    return string(decrypted), nil
}
```

## Security Considerations

1. **Key Storage**
   - Private keys are stored with 0600 permissions
   - Keys are Base64 encoded for storage
   - Backup directory is created with 0700 permissions

2. **Encryption Properties**
   - Provides authenticated encryption
   - Uses unique nonce for each entry
   - Prevents tampering via authentication
   - Forward secrecy: each entry has its own encryption parameters

3. **Memory Security**
   - Keys are zeroed after use
   - Sensitive data is cleared from memory when possible

## Error Handling

All cryptographic operations include comprehensive error handling:
- Key generation failures
- Encryption/decryption errors
- File system errors
- Permission issues
- Invalid key format

## Testing

The cryptographic implementation includes:
1. Unit tests for all crypto operations
2. Integration tests for key management
3. Property-based tests for encryption/decryption
4. Test vectors for known inputs/outputs

## Example Usage

```go
// Generate new keys for a journal
keyID, err := crypto.GenerateKey()
if err != nil {
    log.Fatal(err)
}

// Encrypt an entry
encrypted, err := crypto.Encrypt("My secret journal entry", publicKey)
if err != nil {
    log.Fatal(err)
}

// Decrypt an entry
decrypted, err := crypto.Decrypt(encrypted, publicKey, privateKey)
if err != nil {
    log.Fatal(err)
}
```

## Performance Considerations

- Key generation: One-time operation per journal
- Encryption: O(n) where n is entry length
- Decryption: O(n) where n is entry length
- Memory usage: O(n) for encryption/decryption operations

## Future Considerations

1. **Key Rotation**
   - Implement key rotation capability
   - Maintain key version history
   - Automatic re-encryption with new keys

2. **Backup Enhancement**
   - Encrypted key backup
   - Cloud backup support
   - Key recovery mechanisms

3. **Hardware Security**
   - HSM support
   - TPM integration
   - Secure enclave usage where available 