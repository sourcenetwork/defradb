#!/usr/bin/env bash
set -euo pipefail

export HOME="$(mktemp -d)"

# Run the wizard using expect
expect <<'EOF'
set timeout 15

# Force a TTY via script
spawn script -q -c "./build/defradb wizard" /dev/null

# --- Step 1: Opting in to wizard ---
expect {
    -re ".*setup wizard.*" {}
    timeout { puts "Timeout waiting for welcome screen"; exit 1 }
}
send "\r"

# --- Step 2: Config.yaml generation prompt ---
expect {
    -re ".*config.yaml.*" {}
    timeout { puts "Timeout waiting for config.yaml prompt"; exit 1 }
}
send "\r"

# --- Step 3: Config.yaml generated confirmation ---
expect {
    -re ".*config.yaml.*" {}
    timeout { puts "Timeout waiting for config.yaml confirmation"; exit 1 }
}
send "\r"

# --- Step 4: Keypair explanation ---
expect {
    -re ".*DefraDB.*" {}
    timeout { puts "Timeout waiting for keypair explanation"; exit 1 }
}
send "\r"

# --- Step 5: DEFRA_KEYRING_SECRET requirement ---
expect {
    -re ".*DEFRA_KEYRING_SECRET.*" {}
    timeout { puts "Timeout waiting for DEFRA_KEYRING_SECRET requirement"; exit 1 }
}
send "\r"

# --- Step 6: Enter DEFRA_KEYRING_SECRET value ---
expect {
    -re ".*DEFRA_KEYRING_SECRET.*" {}
    timeout { puts "Timeout waiting for DEFRA_KEYRING_SECRET input"; exit 1 }
}
send "secret-password\r"

# --- Step 7: DEFRA_KEYRING_SECRET confirmation ---
expect {
    -re ".*DEFRA_KEYRING_SECRET.*" {}
    timeout { puts "Timeout waiting for DEFRA_KEYRING_SECRET confirmation"; exit 1 }
}
send "\r"

# --- Step 8: Import existing keys prompt ---
expect {
    -re ".*import.*" {}
    timeout { puts "Timeout waiting for import keys prompt"; exit 1 }
}
send "j\r"

# --- Step 9: Identity key generation prompt ---
expect {
    -re ".*identity.*" {}
    timeout { puts "Timeout waiting for identity key prompt"; exit 1 }
}
send "\r"

# --- Step 10: Keyring file generated confirmation ---
expect {
    -re ".*generated.*" {}
    timeout { puts "Timeout waiting for keyring file confirmation"; exit 1 }
}
send "\r"

# --- Step 11: Test DefraDB configuration ---
expect {
    -re ".*DefraDB.*" {}
    timeout { puts "Timeout waiting for configuration test prompt"; exit 1 }
}
send "\r"

# --- Step 12: Health check ---
expect {
    -re ".*health.*" {}
    timeout { puts "Timeout waiting for health check prompt"; exit 1 }
}
send "\r"

# --- Step 13: Health check completion ---
expect {
    -re ".*ready.*" {}
    timeout { puts "Timeout waiting for health check completion"; exit 1 }
}
send "\r"

# --- Step 14: Wizard completion ---
expect {
    -re ".*complete.*" {}
    timeout { puts "Timeout waiting for wizard completion"; exit 1 }
}
send "\r"


# Wait for process to finish
expect eof
catch wait result
set exit_status [lindex $result 3]
if { $exit_status != 0 } {
    puts "Wizard exited with code $exit_status"
    exit 1
}

puts "Wizard completed successfully"
EOF
