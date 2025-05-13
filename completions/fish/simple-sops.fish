# Completions for simple-sops
# Save this file to ~/.config/fish/completions/simple-sops.fish

# Define all the subcommands
set -l commands encrypt decrypt edit set-keys config rm clean-config get-key clear-key gen-key run help completion

# Define files that can be encrypted/decrypted (matching supported file types)
function __fish_simple_sops_files
    # Use extended globs to find all relevant files
    for ext in yaml yml json ini env properties toml hcl tfvars tfstate pem crt key
        find . -type f -name "*.$ext" 2>/dev/null
    end
end

# Define files that are already encrypted
function __fish_simple_sops_encrypted_files
    # Check each potential file for encryption markers
    for file in (__fish_simple_sops_files)
        if grep -q -e "sops:" -e "\[sops\]" -e "ENC\[AES256_GCM" -e sops_ -e encrypted_suffix "$file" 2>/dev/null
            echo "$file"
        end
    end
end

# Complete subcommands
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a encrypt -d "Encrypt files with Age"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a decrypt -d "Decrypt files"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a edit -d "Edit an encrypted file"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a set-keys -d "Choose which keys to encrypt"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a config -d "Show current SOPS configurations"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a rm -d "Remove files and configurations"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a clean-config -d "Clean orphaned rules"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a get-key -d "Load SOPS key from 1Password"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a clear-key -d "Remove SOPS key"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a gen-key -d "Generate a new Age key pair"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a run -d "Run a command with a decrypted file"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a completion -d "Generate shell completion scripts"
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands" -a help -d "Show help message"

# File completions for the shorthand method (when no command is given, assume it's edit)
complete -c simple-sops -f -n "not __fish_seen_subcommand_from $commands && count (commandline -opc) = 1" -a "(__fish_simple_sops_encrypted_files)"

# Global options
complete -c simple-sops -s q -l quiet -d "Minimal output"
complete -c simple-sops -s d -l debug -d "Show debug information"
complete -c simple-sops -s k -l key-file -d "Age key file to use"

# Complete file arguments for encrypt (use non-encrypted files)
complete -c simple-sops -f -n "__fish_seen_subcommand_from encrypt" -a "(__fish_simple_sops_files)"

# Complete file arguments for decrypt
complete -c simple-sops -f -n "__fish_seen_subcommand_from decrypt" -a "(__fish_simple_sops_encrypted_files)"
complete -c simple-sops -f -n "__fish_seen_subcommand_from decrypt" -a --stdout -d "Output to stdout"

# Complete file arguments for edit
complete -c simple-sops -f -n "__fish_seen_subcommand_from edit" -a "(__fish_simple_sops_encrypted_files)"

# Complete file arguments for set-keys (any yaml/json/ini files)
complete -c simple-sops -f -n "__fish_seen_subcommand_from set-keys" -a "(__fish_simple_sops_files)"

# Complete file arguments for rm (any yaml/json/ini files)
complete -c simple-sops -f -n "__fish_seen_subcommand_from rm" -a "(__fish_simple_sops_files)"

# Complete file arguments for run
complete -c simple-sops -f -n "__fish_seen_subcommand_from run && count (commandline -opc) = 2" -a "(__fish_simple_sops_encrypted_files)"
complete -c simple-sops -f -n "__fish_seen_subcommand_from run && count (commandline -opc) = 3" -a "(__fish_complete_command)"

# Complete gen-key arguments
complete -c simple-sops -f -n "__fish_seen_subcommand_from gen-key" -l force -s f -d "Overwrite existing key file"

# Complete completion subcommand
complete -c simple-sops -f -n "__fish_seen_subcommand_from completion" -a "bash zsh fish powershell" -d "Shell type"

# No arguments for config, clean-config, get-key, clear-key, or help
complete -c simple-sops -f -n "__fish_seen_subcommand_from config clean-config get-key clear-key help"
