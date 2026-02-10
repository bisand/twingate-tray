# Twingate Tray - Installation Guide

## Installation Options

Choose the installation method that best fits your distribution:

### Option 1: DEB Package (Debian-based)

**Recommended for:** Ubuntu, Debian, Zorin OS, Pop!_OS, Linux Mint, Elementary OS

```bash
# Download the appropriate package for your architecture
# AMD64 (most common):
wget https://github.com/bisand/twingate-tray/releases/latest/download/twingate-tray_VERSION_amd64.deb

# ARM64:
wget https://github.com/bisand/twingate-tray/releases/latest/download/twingate-tray_VERSION_arm64.deb

# Install the package
sudo dpkg -i twingate-tray_VERSION_amd64.deb

# If dependencies are missing, run:
sudo apt-get install -f
```

**Dependencies installed:** zenity, libnotify-bin, policykit-1

### Option 2: RPM Package (RPM-based)

**Recommended for:** Fedora, RHEL, CentOS, Rocky Linux, AlmaLinux, openSUSE

```bash
# Download the appropriate package for your architecture
# x86_64 (most common):
wget https://github.com/bisand/twingate-tray/releases/latest/download/twingate-tray-VERSION-1.x86_64.rpm

# ARM64:
wget https://github.com/bisand/twingate-tray/releases/latest/download/twingate-tray-VERSION-1.aarch64.rpm

# Install the package
sudo rpm -i twingate-tray-VERSION-1.x86_64.rpm

# Or using dnf:
sudo dnf install twingate-tray-VERSION-1.x86_64.rpm
```

**Dependencies installed:** zenity, libnotify, polkit

### Option 3: Tarball (Any Linux Distribution)

**Recommended for:** All other distributions, or if you prefer manual installation

```bash
# Download the appropriate tarball for your architecture
# AMD64 (most common):
wget https://github.com/bisand/twingate-tray/releases/latest/download/twingate-tray-linux-amd64.tar.gz

# Extract the tarball
tar -xzf twingate-tray-linux-amd64.tar.gz
cd twingate-tray

# Run the installation script (interactive)
./install.sh

# Or manually copy the binary
sudo cp twingate-tray /usr/local/bin/
sudo chmod +x /usr/local/bin/twingate-tray
```

**Manual dependency installation:**

For Debian/Ubuntu:
```bash
sudo apt install zenity libnotify-bin policykit-1
```

For Fedora/RHEL:
```bash
sudo dnf install zenity libnotify polkit
```

For Arch:
```bash
sudo pacman -S zenity libnotify polkit
```

## Architecture Detection

To find your system architecture:

```bash
uname -m
```

**Architecture mapping:**
- `x86_64` → Download `amd64` or `x86_64` package
- `aarch64` → Download `arm64` or `aarch64` package
- `i686` or `i386` → Download `386` or `i686` package
- `armv7l` → Download `armv7` or `armv7hl` package

## Starting the Application

### Start Manually

```bash
twingate-tray
```

### Start with Systemd (Recommended)

```bash
# Enable and start the service
systemctl --user enable --now twingate-tray

# Check status
systemctl --user status twingate-tray

# View logs
journalctl --user -u twingate-tray -f
```

### Start on Login (Desktop Autostart)

Create `~/.config/autostart/twingate-tray.desktop`:

```ini
[Desktop Entry]
Type=Application
Name=Twingate Tray
Comment=System tray indicator for Twingate VPN
Exec=/usr/local/bin/twingate-tray
Icon=network-vpn
Terminal=false
Categories=Network;
StartupNotify=false
X-GNOME-Autostart-enabled=true
```

## GNOME Desktop Requirements

If you're using GNOME Shell, you need the AppIndicator extension:

```bash
# Ubuntu/Debian
sudo apt install gnome-shell-extension-appindicator

# Fedora
sudo dnf install gnome-shell-extension-appindicator

# Then enable it
gnome-extensions enable appindicatorsupport@ubuntu.com
```

## Verification

After installation, verify the tray icon appears:

1. Start the application (see above)
2. Look for the tray icon in your system tray (usually top-right corner)
3. Right-click the icon to access the menu

## Uninstallation

### DEB Package

```bash
sudo dpkg -r twingate-tray
```

### RPM Package

```bash
sudo rpm -e twingate-tray
```

### Manual Installation

```bash
# Stop the service if running
systemctl --user stop twingate-tray
systemctl --user disable twingate-tray

# Remove files
sudo rm /usr/local/bin/twingate-tray
rm ~/.config/systemd/user/twingate-tray.service
rm ~/.config/autostart/twingate-tray.desktop

# Reload systemd
systemctl --user daemon-reload
```

## Troubleshooting

### Icon not appearing in system tray

**GNOME users:** Ensure AppIndicator extension is installed and enabled (see above)

**Check D-Bus registration:**
```bash
busctl --user list | grep StatusNotifierItem
```

### Permission errors when connecting/disconnecting

Ensure you have either `pkexec` or `sudo` available:
```bash
which pkexec
which sudo
```

### Missing dependencies

**Check for zenity:**
```bash
which zenity
```

**Check for notify-send:**
```bash
which notify-send
```

### Build from source (Advanced)

If you need to build from source:

```bash
# Install build dependencies
# Debian/Ubuntu:
sudo apt install golang-go libx11-dev build-essential

# Fedora:
sudo dnf install golang libX11-devel make

# Clone repository
git clone https://github.com/bisand/twingate-tray.git
cd twingate-tray

# Build
make build

# Install
sudo make install
```

## Getting Help

- **Documentation:** See `README.md` and `USAGE_GUIDE.md`
- **Issues:** https://github.com/bisand/twingate-tray/issues
- **Logs:** Run `twingate-tray` in a terminal to see debug output

## Package Locations

After installation, files are located at:

- **Binary:** `/usr/local/bin/twingate-tray`
- **Systemd service:** `/lib/systemd/user/twingate-tray.service` or `~/.config/systemd/user/twingate-tray.service`
- **Documentation:** `/usr/share/doc/twingate-tray/`
