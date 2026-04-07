#!/bin/sh
# postinstall.sh — Linux .deb/.rpm post-install script.
# Registers /usr/local/bin in /etc/profile.d/simple-cli.sh for all users.
# Called by the package manager after binary extraction.
set -e
cat > /etc/profile.d/simple-cli.sh << 'EOF'
# simple-cli PATH registration (added by package installer)
export PATH="$PATH:/usr/local/bin"
EOF
chmod 644 /etc/profile.d/simple-cli.sh
