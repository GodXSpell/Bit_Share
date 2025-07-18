# BitShare P2P Mesh Network

A high-speed peer-to-peer file sharing system with mesh network capabilities.

## Installation Options

### Option 1: Direct Download (Recommended for most users)

Download the latest pre-compiled binary for your system:

- [Download BitShare for Windows](https://github.com/yourusername/bitshare/releases/latest/download/bitshare-windows.zip)
- [Download BitShare for macOS](https://github.com/yourusername/bitshare/releases/latest/download/bitshare-macos.zip)
- [Download BitShare for Linux](https://github.com/yourusername/bitshare/releases/latest/download/bitshare-linux.tar.gz)

After downloading, extract the archive and run the installer:
- Windows: Run `install.bat` as administrator
- macOS/Linux: Run `./install.sh`

### Option 2: Using a Package Manager

```bash
# Using Homebrew (macOS and Linux)
brew tap yourusername/bitshare
brew install bitshare

# Using Scoop (Windows)
scoop bucket add bitshare https://github.com/yourusername/scoop-bitshare.git
scoop install bitshare
```

### Option 3: From Source (For developers)

If you prefer to build from source, follow these steps:

#### Windows
```powershell
cd d:\Bit_Share\go
go build -o bitshare.exe
```

#### Linux/macOS
```bash
cd /path/to/Bit_Share/go
go build -o bitshare
```

## Automatic Updates

BitShare checks for updates when started. When an update is available, you can:

```
# Check for updates manually
bitshare update check

# Install the latest update
bitshare update install

# Enable automatic updates (runs on startup)
bitshare update auto --enable
```

## Usage

Start interactive mode:
```
bitshare
```

Or use specific commands:
```
bitshare scan       # Scan for peers
bitshare list       # List known peers
bitshare receive 9000 C:\Downloads    # Receive files
bitshare send laptop-name 9000 file.pdf    # Send a file
```

## Features

- Direct P2P connections using WiFi Direct, TCP/IP, and Bluetooth
- Mesh network for extended range through intermediate nodes
- High-performance chunked file transfers with parallel streams
- Resumable downloads for reliability
- Works on networks with client isolation (hotels, cafes, etc.)

## Requirements

- Go 1.16 or newer
