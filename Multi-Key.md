# Multiple Key Support

One of the key features in this version of simple-sops is the ability to use multiple Age keys for encryption. This allows different people or systems to decrypt the same files with their respective keys.

## Encrypting with Multiple Keys

You can use multiple keys in several ways:

### 1. Multiple Key Files

```bash
# Encrypt with multiple key files separated by commas
simple-sops encrypt --key-files ~/.key1.txt,~/.key2.txt secret.yaml
```

This will combine the keys from all specified files and use them for encryption. Each person with any of these keys will be able to decrypt the file.

### 2. Multiple Keys from 1Password

```bash
# Encrypt with keys stored in different 1Password items
simple-sops encrypt --op-items "MY_KEY,TEAM_KEY" secret.yaml

# Specify different vaults for each key
simple-sops encrypt --op-items "MY_KEY,TEAM_KEY" --op-vaults "Personal,Work" secret.yaml

# Specify a custom field name
simple-sops encrypt --op-items "MY_KEY" --op-field "private_key" secret.yaml
```

This allows you to retrieve keys from multiple 1Password items, possibly across different vaults, and use them all for encryption.

## Under the Hood

When encrypting with multiple keys, SOPS:

1. Generates a single data key to encrypt the actual content
2. Encrypts this data key separately with each of the public keys you provide

When decrypting, only one of the keys is needed - SOPS will try each key to see if it can decrypt the data key. As long as you possess any of the keys used for encryption, you'll be able to decrypt the file.

## SOPS Configuration

When using multiple keys, simple-sops updates the `.sops.yaml` file to include all keys. For example:

```yaml
creation_rules:
  - path_regex: secret.yaml
    age: age1abcdef123456,age1xyz789abcdef
```

This ensures that subsequent encryption operations use all the specified keys.

## Example Use Cases

1. **Team Access**: Different team members can decrypt the same files using their own keys
2. **Environment Access**: Access the same secrets from different environments (e.g., your laptop and a Kubernetes cluster)
3. **Backup Keys**: Have backup keys that can decrypt critical files

These features make simple-sops a powerful tool for managing encrypted configuration files in team and multi-environment settings.
