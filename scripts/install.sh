#!/bin/bash
# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
APP_NAME="keyball-gestures" # Consider "char5742-keyball-gestures" if more uniqueness needed
INSTALL_PATH="/usr/local/bin/${APP_NAME}"
UDEV_RULE_FILE="/etc/udev/rules.d/99-${APP_NAME}.rules" # More unique rule file name
GITHUB_REPO="char5742/keyball-gestures"
REQUIRED_GROUP="input" # Standard group for uinput access
# --- Configuration End ---

echo "Starting ${APP_NAME} installation..."
echo "This script requires sudo privileges for installation."

# Prompt for sudo password upfront and keep the session alive
sudo -v
while true; do sudo -n true; sleep 60; kill -0 "$$" || exit; done 2>/dev/null &
SUDO_KEEPALIVE_PID=$!
# Ensure the keepalive process is killed on script exit
trap 'echo "Exiting..."; kill "$SUDO_KEEPALIVE_PID" &>/dev/null; exit' INT TERM EXIT

# 1. Check Architecture and determine binary name
echo "--> Checking system architecture..."
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  BINARY_SUFFIX="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  BINARY_SUFFIX="arm64"
else
  echo "Error: Unsupported architecture: $ARCH" >&2
  # Trap will handle cleanup and exit
  exit 1
fi
BINARY_NAME="${APP_NAME}-${BINARY_SUFFIX}"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}"

# 2. Download the binary to a temporary location
# Create a temporary directory that will be cleaned up automatically by the trap
TMP_DIR=$(mktemp -d)
# Modify trap to ensure TMP_DIR is removed
trap 'echo "Cleaning up..."; rm -rf "$TMP_DIR"; kill "$SUDO_KEEPALIVE_PID" &>/dev/null; exit' INT TERM EXIT

echo "--> Downloading ${BINARY_NAME} from ${DOWNLOAD_URL}..."
wget --progress=bar:force -O "${TMP_DIR}/${APP_NAME}" "${DOWNLOAD_URL}"
echo "Download complete."

# 3. Make the binary executable
echo "--> Setting execute permissions..."
chmod +x "${TMP_DIR}/${APP_NAME}"

# 4. Move the binary to the installation path
echo "--> Installing ${APP_NAME} to ${INSTALL_PATH}..."
# Check if already exists and ask (optional, currently overwrites)
if [ -f "${INSTALL_PATH}" ]; then
    echo "Warning: ${INSTALL_PATH} already exists. Overwriting."
fi
sudo mv "${TMP_DIR}/${APP_NAME}" "${INSTALL_PATH}"

# 5. Create udev rule
echo "--> Creating udev rule at ${UDEV_RULE_FILE}..."
UDEV_RULE_CONTENT="KERNEL==\"uinput\", MODE=\"0660\", GROUP=\"${REQUIRED_GROUP}\""
# Use tee to write the file with sudo privileges
echo "${UDEV_RULE_CONTENT}" | sudo tee "${UDEV_RULE_FILE}" > /dev/null
echo "udev rule created."

# 6. Add user to the required group if not already a member
NEEDS_RELOGIN="false"
if ! groups "$USER" | grep -q "\b${REQUIRED_GROUP}\b"; then
    echo "--> Adding user '${USER}' to the '${REQUIRED_GROUP}' group for uinput access..."
    sudo usermod -aG "${REQUIRED_GROUP}" "$USER"
    echo "User added to the group."
    NEEDS_RELOGIN="true"
else
    echo "--> User '${USER}' is already a member of the '${REQUIRED_GROUP}' group."
fi

# 7. Reload udev rules
echo "--> Reloading udev rules..."
sudo udevadm control --reload-rules
sudo udevadm trigger --subsystem-match=misc # Trigger uinput device events

# 8. Clean up (TMP_DIR is cleaned by trap)

# Kill the sudo keepalive process explicitly before final message
kill "$SUDO_KEEPALIVE_PID" &>/dev/null
# Disable the trap on a clean exit, otherwise it might echo "Cleaning up..." unnecessarily
trap - INT TERM EXIT

# 9. Final instructions
echo ""
echo "--------------------------------------------------"
echo "${APP_NAME} installation successful!"
echo "--------------------------------------------------"
if [ "$NEEDS_RELOGIN" = "true" ]; then
    echo "IMPORTANT: You MUST log out and log back in for the group changes to take effect."
    echo "After logging back in, you can run the application with:"
else
    echo "You can now run the application with:"
fi
echo "  ${APP_NAME} --gui"
echo ""
echo "To uninstall ${APP_NAME}, manually run:"
echo "  sudo rm ${INSTALL_PATH}"
echo "  sudo rm ${UDEV_RULE_FILE}"
echo "  sudo udevadm control --reload-rules"
echo "  # Optionally, remove user from group: sudo gpasswd -d \$USER ${REQUIRED_GROUP}"
echo "--------------------------------------------------"

exit 0