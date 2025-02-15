# JOT
```
    __                                  
   |  |                    _     
 __|  |___ _ _ ___ ___ ___| |    
|  |  | . | | |  _|   | .'| |    
|_____|___|___|_| |_|_|__,|_|    
                                 
                                 
 _____        _   _              
|     |___   | |_| |_ ___        
|  |  |   |  |  _|   | -_|       
|_____|_|_|  |_| |_|_|___|       
                                 
                                 
 _____               _         _ 
|_   _|___ ___ _____|_|___ ___| |
  | | | -_|  _|     | |   | .'| |
  |_| |___|_| |_|_|_|_|_|_|__,|_|
                                 
```

JOT is a command-line journaling tool that allows you to create and manage encrypted journals.

## Features

- Create and manage multiple journals
- Automatic GPG encryption for journal security
- NaCl encryption for secure journal storage
- Default journal support for quick entries
- Simple command-line interface

## Installation

```bash
go install github.com/veritome/jot/cmd/jot@latest
```

## Usage

### Journal Management

```bash
# Create a new journal
jot journal new <name>

# Set default journal
jot journal default <name>

# List entries in a journal
jot journal read <name>

# Show journal information
jot journal describe <name>

# Delete a journal
jot journal delete <name>
```

### Creating Entries

```bash
# Add entry to default journal
jot "Your journal entry text here"

# Add entry to specific journal
jot --journal <name> "Your journal entry text here"
```

## Storage

All journal data is stored securely in `$HOME/.jot/` directory. 