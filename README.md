# Jotta Archiver

A terminal user interface (TUI) tool for archiving folders using jotta-cli with configurable presets and progress monitoring.

## Features

- 📦 **Interactive TUI** - Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a smooth terminal experience
- ⚙️ **Configurable Presets** - Define multiple archive destinations in a YAML config file
- 🔧 **Custom Remote Paths** - Enter custom remote directories on-the-fly without editing config
- 🎲 **Auto-generated Names** - Archives are automatically named with format `YYYYMMDD_random_word_combo`
- ✏️ **Editable Names** - Customize the archive name before uploading
- 📁 **Format Subfolders** - Optionally split files into per-format subfolders (e.g. `JPEG/`, `HEIF/`) per preset
- 🔀 **Extension Renaming** - Remap file extensions before upload (e.g. `.HIF` → `.HEIF`) without converting files
- 📊 **Progress Monitoring** - Seamlessly launches jotta-cli's built-in observe TUI
- 🔄 **Background Upload** - Exit the observer and uploads continue in the background
- 🐛 **Debug Mode** - Toggle debug mode to see jotta-cli commands and output
- 🖱️ **macOS Finder Integration** - Right-click folders in Finder to launch archiver (Quick Action)

## Prerequisites

- [jotta-cli](https://docs.jottacloud.com/en/articles/1437201-jotta-cli-documentation) must be installed and configured
- Go 1.21 or later (for building from source)

## Installation

### Build from Source

```bash
git clone https://github.com/luca/jotta-archiver.git
cd jotta-archiver
go build -o jotta-archiver
sudo mv jotta-archiver /usr/local/bin/  # Optional: install globally
```

### Verify jotta-cli Installation

```bash
jotta-cli --version
```

## macOS Finder Integration

### Quick Action (Right-Click Menu)

You can integrate jotta-archiver with macOS Finder to archive folders directly from the context menu.

**Installation:**

```bash
cd jotta-archiver
./install/install-finder-integration.sh
```

This will:
- Create an Automator Quick Action workflow
- Install it to `~/Library/Services/`
- Install the binary to `/usr/local/bin/` or `~/.local/bin/`

**Usage:**

1. Right-click (or Control+click) on any folder in Finder
2. Navigate to **Quick Actions** or **Services** in the context menu
3. Click **Archive with Jotta**
4. Terminal will open with jotta-archiver running for that folder
5. Follow the TUI prompts to complete the archive

**Note:** If the Quick Action doesn't appear immediately, you may need to log out and log back in, or run:
```bash
/System/Library/CoreServices/pbs -flush
```

**Uninstall:**
```bash
rm -rf ~/Library/Services/"Archive with Jotta.workflow"
```

## Configuration

On first run, a default configuration file will be created at `~/.jotta-archiver.yaml`:

```yaml
presets:
  - name: "Camera Pictures"
    remote: "/media/pictures/camera_pictures"
    split_by_format: true
  - name: "Documents"
    remote: "/media/documents"
  - name: "Music"
    remote: "/media/music/archives"

extension_renames:
  HIF: HEIF
```

### Config File Format

Edit `~/.jotta-archiver.yaml` to add your own presets:

```yaml
presets:
  - name: "Your Preset Name"
    remote: "/your/remote/path"
    split_by_format: true   # optional
  - name: "Another Preset"
    remote: "/another/remote/path"

extension_renames:          # optional, global
  HIF: HEIF
  heic: heic
```

Each preset requires:
- `name`: A descriptive name shown in the TUI
- `remote`: The remote path on Jottacloud where archives will be stored
- `split_by_format` *(optional)*: When `true`, files are grouped into per-format subfolders at upload time (see below)

### Format Subfolders (`split_by_format`)

When a preset has `split_by_format: true`, files are organized into a subfolder named after their (uppercased) extension, inserted as the immediate parent of each file. The rest of the directory hierarchy is preserved:

```
Source folder:           Uploaded as:
vacation/                vacation/
  IMG_001.jpg              JPEG/
  IMG_001.HIF     →          IMG_001.jpg
  clip.mp4               HEIF/
                             IMG_001.HEIF   ← renamed from .HIF
                         MP4/
                             clip.mp4
```

No files are copied or converted — hard links are used so preprocessing is near-instantaneous. The original folder is never modified.

### Extension Renaming (`extension_renames`)

The global `extension_renames` map renames file extensions before upload without converting the file content. Keys are matched case-insensitively:

```yaml
extension_renames:
  HIF: HEIF     # .HIF, .hif, .Hif → .HEIF
```

This is useful for formats where the extension varies but the container is the same (e.g. Apple's `.HIF` files are HEIF images and some software expects the `.HEIF` extension).

## Usage

```bash
jotta-archiver <folder>
```

### Example

```bash
jotta-archiver ~/Pictures/vacation_2025
```

### Interactive Flow

1. **Select Preset**: Use arrow keys (↑/↓) to navigate presets, or select "Custom..." for manual remote path entry, press Enter to select
2. **Custom Remote** (if selected): Enter a custom remote directory path
3. **Edit Archive Name**: Modify the auto-generated name if desired, press Enter to start upload
4. **Monitor Progress**: Automatically launches `jotta-cli observe` which provides a real-time TUI for monitoring upload progress
5. **Debug Mode**: Press 'd' to toggle debug mode and see command output before starting upload

### Keyboard Controls

#### Preset Selection Screen
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - Select preset or custom option
- `d` - Toggle debug mode
- `q` / `Ctrl+C` - Quit

#### Custom Remote Screen
- Type to enter custom remote path
- `Enter` - Continue to name editing
- `Esc` - Go back to preset selection
- `Ctrl+C` - Quit

#### Name Editing Screen
- Type to edit the archive name
- `Enter` - Start upload (launches jotta-cli observe)
- `Esc` - Go back to previous screen
- `d` - Toggle debug mode
- `Ctrl+C` - Quit

## How It Works

1. **Archive Name Generation**: Creates a unique name using the format `YYYYMMDD_word1_word2_word3` (e.g., `20251029_swift_mountain_river`)

2. **Preprocessing** *(if enabled)*: Before uploading, a temporary directory is built using hard links (zero-copy) with files reorganized into format subfolders and/or with extensions renamed according to config. The original folder is never modified.

3. **Archive Command**: Executes:
   ```bash
   jotta-cli archive <folder> --remote=<remote_path>/<archive_name>
   ```
   Captures the upload ID from the output

4. **Progress Monitoring**: Automatically launches jotta-cli's built-in observer:
   ```bash
   jotta-cli observe --uploadid=<upload_id>
   ```
   This provides a real-time TUI showing upload progress, speed, and status

5. **Cleanup**: The temporary preprocessing directory is removed after the upload completes.

6. **Debug Mode**: When enabled (press 'd'), displays:
   - Commands being executed
   - Raw output from jotta-cli archive
   - Upload ID that was captured
   - Useful for troubleshooting upload issues

## Development

### Project Structure

```
jotta-archiver/
├── main.go                 # Entry point, CLI parsing, orchestration
├── config/
│   └── config.go          # YAML config loading
├── wordgen/
│   └── wordgen.go         # Random archive name generation
├── preprocess/
│   └── preprocess.go      # Format splitting and extension renaming
├── archive/
│   └── archive.go         # Jotta-cli command execution
├── tui/
│   └── tui.go             # Bubble Tea TUI implementation
└── go.mod                 # Go dependencies
```

### Dependencies

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

### Building

```bash
go build -o jotta-archiver
```

### Running Tests

```bash
go test ./...
```

## Troubleshooting

### "jotta-cli not found"
Ensure jotta-cli is installed and in your PATH:
```bash
which jotta-cli
```

### "failed to load config"
Check that `~/.jotta-archiver.yaml` exists and is valid YAML. The tool will create a default config on first run if none exists.

### "folder does not exist"
Verify the folder path is correct. Use absolute paths if relative paths aren't working:
```bash
jotta-archiver /full/path/to/folder
```

### Upload not starting
If the upload doesn't start after pressing Enter:
- Check that jotta-cli is properly configured (`jotta-cli status`)
- Verify the remote path exists or has proper permissions
- Enable debug mode (`d` key) to see the exact commands and errors

### Debugging upload issues
Press `d` to toggle debug mode before starting upload to see:
- The exact jotta-cli archive command being executed
- Raw output from jotta-cli including the upload ID
- Any error messages from the command
- Useful for diagnosing path or permission issues

## License

MIT License - feel free to use and modify as needed.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
