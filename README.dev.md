# JOT Development Documentation

## Project Structure

```
cmd/jot/           # Main CLI application
internal/
  journal/         # Journal management
  entry/           # Entry management
  crypto/          # Encryption utilities
docs/              # Additional documentation
```

## Design Principles

1. **Standard Library Only**: No external dependencies except Go standard library
2. **Security First**: All journal data is encrypted using NaCl for modern security
3. **Simple Interface**: Clear and intuitive CLI commands
4. **Data Storage**: All data stored in `$HOME/.jot/` directory

## Core Components

### Journal

- Unique name identifier
- Creation timestamp
- List of entries
- NaCl key for encryption

### Entry

- Creation timestamp
- Encrypted body text
- Associated with a specific journal

## Development Guidelines

1. Follow Go 1.20+ standards
2. Adhere to:
   - SOLID principles
   - DRY (Don't Repeat Yourself)
   - KISS (Keep It Simple, Stupid)
   - YAGNI (You Aren't Gonna Need It)
3. Document all exported functions and types
4. Keep files focused and concise
5. Implement proper error handling
6. Use the principle of least privilege

## Testing

- Write unit tests for all packages
- Include integration tests for CLI commands
- Test encryption/decryption functionality thoroughly

## Building

```bash
go build ./cmd/jot
``` 