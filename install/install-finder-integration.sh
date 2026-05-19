#!/bin/bash
# Installation script for jotta-archiver Finder integration

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
WORKFLOW_NAME="Archive with Jotta.workflow"
SERVICES_DIR="$HOME/Library/Services"
BINARY_NAME="jotta-archiver"

echo "🚀 Installing jotta-archiver Finder Integration"
echo "================================================"
echo ""

# Check if binary exists
if [ ! -f "$PROJECT_ROOT/$BINARY_NAME" ]; then
    echo "❌ Error: $BINARY_NAME binary not found in $PROJECT_ROOT"
    echo "Please build the project first: go build -o jotta-archiver"
    exit 1
fi

# Create workflow if it doesn't exist
if [ ! -d "$SCRIPT_DIR/$WORKFLOW_NAME" ]; then
    echo "📦 Creating Automator workflow..."
    bash "$SCRIPT_DIR/create-workflow.sh"
    echo ""
fi

# Create Services directory if it doesn't exist
mkdir -p "$SERVICES_DIR"

# Copy workflow to Services
echo "📋 Copying workflow to $SERVICES_DIR..."
rm -rf "$SERVICES_DIR/$WORKFLOW_NAME"
cp -R "$SCRIPT_DIR/$WORKFLOW_NAME" "$SERVICES_DIR/"
echo "✅ Workflow installed"
echo ""

# Install binary to a standard location
echo "🔧 Installing binary..."

# Check if /usr/local/bin is writable
if [ -w "/usr/local/bin" ]; then
    INSTALL_PATH="/usr/local/bin/$BINARY_NAME"
    cp "$PROJECT_ROOT/$BINARY_NAME" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
    echo "✅ Binary installed to $INSTALL_PATH"
elif [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    INSTALL_PATH="$HOME/.local/bin/$BINARY_NAME"
    cp "$PROJECT_ROOT/$BINARY_NAME" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
    echo "✅ Binary installed to $INSTALL_PATH"
    
    # Check if ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo ""
        echo "⚠️  Warning: $HOME/.local/bin is not in your PATH"
        echo "Add this line to your ~/.zshrc or ~/.bash_profile:"
        echo ""
        echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
    fi
else
    echo ""
    echo "⚠️  Could not install to /usr/local/bin (not writable)"
    echo "Please run: sudo cp $PROJECT_ROOT/$BINARY_NAME /usr/local/bin/"
    echo "Or add it to your PATH manually"
    INSTALL_PATH="$PROJECT_ROOT/$BINARY_NAME"
fi

echo ""
echo "✨ Installation Complete!"
echo ""
echo "📖 How to use:"
echo "1. Right-click (or Control+click) on any folder in Finder"
echo "2. Navigate to 'Quick Actions' or 'Services' in the context menu"
echo "3. Click 'Archive with Jotta'"
echo "4. Terminal will open with jotta-archiver running on that folder"
echo ""
echo "Note: If you don't see the Quick Action immediately, you may need to:"
echo "- Log out and log back in, or"
echo "- Run: /System/Library/CoreServices/pbs -flush"
echo ""
echo "To uninstall, simply delete:"
echo "  $SERVICES_DIR/$WORKFLOW_NAME"
echo ""

