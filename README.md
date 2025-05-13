# Simple SOPS

A user-friendly wrapper for Mozilla SOPS (Secrets OPerationS) with Age encryption, providing an easier way to manage encrypted configuration files.

## Features

- Simple command-line interface for encrypting and decrypting files
- Seamless integration with 1Password for secure key storage
- Git-aware configuration management
- Age key generation and management
- Selective encryption with customizable patterns
- Multiple file type support
- Run command support for working with encrypted environment files

## Installation

### Prerequisites

- [SOPS](https://github.com/mozilla/sops) - `brew install sops` (macOS) or equivalent
- [Age](https://github.com/FiloSottile/age) - `brew install age` (macOS) or equivalent
- [1Password CLI](https://developer.1password.com/docs/cli/get-started) (optional, for key storage in 1Password)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/simple-sops
cd simple-sops

# Build the binary
go build -o simple-sops ./cmd/simple-sops

# Install to a directory in your PATH (optional)
sudo mv simple-sops /usr/local/bin/

# Generate Fish shell completions (if you use Fish)
mkdir -p ~/.config/fish/completions
simple-sops completion fish > ~/.config/fish/completions/simple-sops.fish
```

## Quick Start

```bash
# Generate a new Age key
simple-sops gen-key

# Configure which parts of a file to encrypt
simple-sops set-keys config.yaml

# Encrypt a file
simple-sops encrypt config.yaml

# Decrypt a file
simple-sops decrypt config.yaml

# Edit an encrypted file directly
simple-sops edit config.yaml
```

## Commands Reference

### Key Management

#### `gen-key` - Generate a new Age key pair

Generate a new Age key for use with SOPS.

```bash
# Generate a key with default settings
simple-sops gen-key

# Specify a custom location
simple-sops gen-key --key-file ~/.keys/my-age-key.txt

# Force overwrite an existing key
simple-sops gen-key --force
```

#### `get-key` - Load key from 1Password

Retrieve the Age key from 1Password and store it in a temporary file.

```bash
simple-sops get-key
```

The key will be stored in a temporary file and the `SOPS_AGE_KEY_FILE` environment variable will be set to its path.

#### `clear-key` - Remove temporary key

Clear the Age key retrieved from 1Password.

```bash
simple-sops clear-key
```

### File Operations

#### `encrypt` - Encrypt files

Encrypt one or more files with Age.

```bash
# Encrypt a single file
simple-sops encrypt config.yaml

# Encrypt multiple files
simple-sops encrypt config.yaml secrets.json .env

# Encrypt with a specific key
simple-sops encrypt --key-file ~/.config/simple-sops/key.txt config.yaml
```

#### `decrypt` - Decrypt files

Decrypt one or more encrypted files.

```bash
# Decrypt a file (interactive mode)
simple-sops decrypt config.yaml

# Decrypt to stdout
simple-sops decrypt --stdout config.yaml

# Decrypt to stdout and pipe to another command
simple-sops decrypt --stdout config.yaml | kubectl apply -f -
```

#### `edit` - Edit an encrypted file

Edit an encrypted file directly.

```bash
# Edit a file (will be decrypted, opened in editor, then re-encrypted)
simple-sops edit secrets.yaml

# Shorthand form (simple-sops defaults to edit when given just a file)
simple-sops secrets.yaml
```

#### `set-keys` - Configure encryption patterns

Choose which keys to encrypt in a file.

```bash
simple-sops set-keys config.yaml
```

This will prompt you to select from several predefined patterns:

1. All values (encrypt entire file)
2. Kubernetes (encrypt data, stringData, password, ingress, token fields)
3. Talos configuration (encrypt secrets sections, certs, keys)
4. Common sensitive data (encrypt passwords, tokens, keys, credentials)
5. Custom pattern (provide your own regex)

#### `run` - Run a command with a decrypted file

Decrypt a file temporarily to run a command, then clean up.

```bash
# Run a command with an encrypted file
simple-sops run encrypted-config.yaml "kubectl apply -f"

# Specify both input and output files
simple-sops run encrypted.env decrypted.env "docker-compose --env-file decrypted.env up"
```

### Configuration Management

#### `config` - Show SOPS configuration

Display the current SOPS configuration.

```bash
simple-sops config
```

#### `rm` - Remove files and configurations

Remove files and their SOPS configurations.

```bash
# Remove a file and its SOPS configuration
simple-sops rm secrets.yaml

# Remove multiple files
simple-sops rm config.yaml secrets.json
```

#### `clean-config` - Clean orphaned rules

Remove rules for files that no longer exist from the SOPS configuration.

```bash
simple-sops clean-config
```

### Shell Integration

#### `completion` - Generate shell completions

Generate shell completion scripts for bash, zsh, fish, or powershell.

```bash
# Generate Fish completions
simple-sops completion fish > ~/.config/fish/completions/simple-sops.fish

# Generate Bash completions
simple-sops completion bash > ~/.bash_completion.d/simple-sops
```

## Common Workflows

### Setting up a new project

```bash
# Generate a new Age key if you don't have one
simple-sops gen-key

# Create and add a rule for selectively encrypting your YAML file
simple-sops set-keys config.yaml
# Select option 2 (Kubernetes) for Kubernetes manifests
# Or option 4 (Common sensitive data) for general config files

# Encrypt the file
simple-sops encrypt config.yaml

# Commit both the encrypted file and .sops.yaml to git
git add config.yaml .sops.yaml
git commit -m "Add encrypted configuration"
```

### Storing your Age key in 1Password (recommended)

1. Store your Age key in 1Password:

   - Create a new item named "SOPS_AGE_KEY_FILE" in your Personal vault
   - Add the content of your key file as a text field named "text"

2. Use the key directly from 1Password:

   ```bash
   # Load the key from 1Password
   simple-sops get-key

   # Now you can encrypt/decrypt without specifying a key file
   simple-sops encrypt config.yaml
   simple-sops decrypt config.yaml

   # Clear the key when done
   simple-sops clear-key
   ```

### Working with Kubernetes Secrets

```bash
# Create a Kubernetes secret file
cat > secret.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
stringData:
  username: admin
  password: supersecret
  api-key: 1234567890abcdef
EOF

# Set up selective encryption for Kubernetes secrets
simple-sops set-keys secret.yaml
# Choose option 2 (Kubernetes)

# Encrypt the secret
simple-sops encrypt secret.yaml

# Apply the encrypted secret directly to the cluster
simple-sops decrypt --stdout secret.yaml | kubectl apply -f -

# Or use the run command
simple-sops run secret.yaml "kubectl apply -f secret.yaml"
```

## Troubleshooting

### Common Issues

1. **"Key file not found"**:

   - Ensure the Age key exists at the specified path
   - Try using `simple-sops gen-key` to create a new key
   - If using 1Password, ensure the item is correctly set up

2. **"Failed to encrypt/decrypt file"**:

   - Check if SOPS is installed properly
   - Verify that the correct key is being used
   - For 1Password integration, ensure you're logged in (`op signin`)

3. **"Cannot determine SOPS config path"**:

   - You may be outside a Git repository
   - Try creating a `.sops.yaml` file in your current directory
   - Or use `set-keys` command to create one automatically

4. **"Command not found" or shell completion not working**:
   - Ensure the binary is in your PATH
   - For Fish shell, verify completions are in `~/.config/fish/completions/`

## Environment Variables

- `SOPS_AGE_KEY_FILE`: Path to the Age key file
- `EDITOR`: Editor to use when editing encrypted files

## Credits

- [SOPS](https://github.com/mozilla/sops) - Mozilla's Secrets OPerationS
- [Age](https://github.com/FiloSottile/age) - A simple, modern and secure encryption tool
