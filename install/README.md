# macOS Finder Integration

This directory contains scripts to integrate jotta-archiver with macOS Finder, allowing you to right-click folders and archive them directly.

## Quick Installation

```bash
./install-finder-integration.sh
```

This automated script will:
1. Create the Automator Quick Action workflow
2. Install it to `~/Library/Services/`
3. Install the jotta-archiver binary to a standard location

## Manual Installation

If you prefer to install manually:

### Step 1: Create the Workflow

```bash
./create-workflow.sh
```

This creates the `Archive with Jotta.workflow` bundle.

### Step 2: Install to Services

```bash
./install-finder-integration.sh
```

## How It Works

The Quick Action uses an AppleScript that:
1. Receives the selected folder from Finder
2. Converts it to a POSIX path
3. Searches for jotta-archiver in common locations
4. Opens Terminal and runs `jotta-archiver <folder-path>`

## Usage

After installation:
1. **Right-click** any folder in Finder
2. Navigate to **Quick Actions** → **Archive with Jotta**
3. Terminal opens with the TUI
4. Select your preset and archive name
5. Watch the upload progress

## Troubleshooting

### Quick Action doesn't appear

If the Quick Action doesn't show up in the context menu:

1. **Restart Finder:**
   ```bash
   killall Finder
   ```

2. **Flush the services cache:**
   ```bash
   /System/Library/CoreServices/pbs -flush
   ```

3. **Log out and log back in**

4. **Check System Preferences:**
   - Open **System Preferences** → **Extensions** → **Finder**
   - Ensure "Archive with Jotta" is checked

### Binary not found error

If you get an error that jotta-archiver cannot be found:

1. Verify the binary location:
   ```bash
   which jotta-archiver
   ```

2. If not found, install it:
   ```bash
   sudo cp ../jotta-archiver /usr/local/bin/
   ```

3. Make sure it's executable:
   ```bash
   chmod +x /usr/local/bin/jotta-archiver
   ```

### Terminal doesn't open

If Terminal doesn't open when you click the Quick Action:

1. Check Terminal permissions in **System Preferences** → **Security & Privacy** → **Automation**
2. Ensure the workflow has permission to control Terminal

## Customization

### Using Different Terminal Applications

The workflow uses Terminal.app by default, but you can easily change it to use iTerm2 or Kitty.

#### Using iTerm2

1. Open the workflow in Automator:
   ```bash
   open ~/Library/Services/"Archive with Jotta.workflow"
   ```

2. Replace `Terminal` with `iTerm`:
   ```applescript
   tell application "iTerm"
       activate
       create window with default profile command binaryPath & " " & quoted form of folderPath
   end tell
   ```

#### Using Kitty Terminal

1. Open the workflow in Automator:
   ```bash
   open ~/Library/Services/"Archive with Jotta.workflow"
   ```

2. Replace the Terminal section with:
   ```applescript
   do shell script "open -a kitty.app --args " & binaryPath & " " & quoted form of folderPath
   ```

   Or for better integration:
   ```applescript
   do shell script "/Applications/kitty.app/Contents/MacOS/kitty --single-instance " & binaryPath & " " & quoted form of folderPath & " &> /dev/null &"
   ```

**Note:** Make sure Kitty is installed and accessible at `/Applications/kitty.app/`. Adjust the path if you installed it elsewhere.

### Changing the Icon

The workflow uses the default action icon. To customize:

1. Open the workflow in Automator
2. Go to **Workflow Settings** (gear icon)
3. Drag a custom icon to the image well

## Uninstallation

To remove the Finder integration:

```bash
rm -rf ~/Library/Services/"Archive with Jotta.workflow"
/System/Library/CoreServices/pbs -flush
```

## Technical Details

**Workflow Type:** Quick Action (Service)  
**Input:** Folders in Finder  
**Action:** Run AppleScript  
**Location:** `~/Library/Services/`  

The workflow bundle contains:
- `Contents/Info.plist` - Bundle metadata and service configuration
- `Contents/document.wflow` - Automator workflow definition with AppleScript

## Requirements

- macOS 10.10 or later
- jotta-cli installed and configured
- jotta-archiver binary built
- Terminal application:
  - Terminal.app (default)
  - iTerm2 (customizable)
  - Kitty (customizable)
  - Or any other terminal app

