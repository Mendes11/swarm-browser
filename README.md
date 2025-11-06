# Swarm Browser

A Terminal User Interface (TUI) application for browsing and managing Docker Swarm clusters. Navigate through your Docker Swarm stacks, services, and tasks with an interactive interface, and attach directly to containers - all from your terminal.

## Features

- 🔍 Browse Docker Swarm clusters, stacks, services, and tasks
- 🔗 Connect to remote Docker hosts via SSH tunneling
- 📦 Attach to running containers directly from the TUI
- ⌨️ Keyboard-driven navigation with intuitive controls
- 🎨 Clean, responsive terminal interface built with Bubble Tea

## Installation

### Quick Install (Recommended)

The easiest way to install Swarm Browser is using our installation script:

#### System-wide installation (requires sudo):
```bash
curl -fsSL https://raw.githubusercontent.com/Mendes11/swarm-browser/main/install.sh | sudo bash
```

#### User-local installation:
```bash
curl -fsSL https://raw.githubusercontent.com/Mendes11/swarm-browser/main/install.sh | bash
```

> **Note:** User-local installation installs to `~/.local/bin`. Make sure this directory is in your PATH.

### Manual Installation

You can also download the binary directly from the [releases page](https://github.com/Mendes11/swarm-browser/releases):

1. Download the appropriate archive for your OS and architecture
2. Extract the archive: `tar -xzf swarm-browser_*.tar.gz`
3. Move the binary to your preferred location (e.g., `/usr/local/bin` or `~/.local/bin`)
4. Make it executable: `chmod +x swarm-browser`

### Build from Source

If you have Go 1.24.0+ installed:

```bash
git clone https://github.com/Mendes11/swarm-browser.git
cd swarm-browser
go build -o swarm-browser main.go
```

## Configuration

Swarm Browser requires a configuration file (`clusters.yml`) to define your Docker Swarm clusters:

```yaml
clusters:
  production:
    name: "Production Cluster"
    host: "prod-manager-01"  # Primary manager node
    nodes:
      manager-01:
        host: "prod-01"        # SSH config hostname
        hostname: "prod-01"    # Docker Swarm node hostname
      worker-01:
        host: "prod-02"
        hostname: "prod-02"
```

### Prerequisites

- SSH access to your Docker Swarm nodes
- Docker Swarm nodes configured in your `~/.ssh/config`
- Docker socket access on remote hosts

Example SSH config entry:

```ssh-config
Host prod-01
    HostName 10.0.1.10
    User docker
    IdentityFile ~/.ssh/id_rsa
```

## Usage

Run the application:

```bash
swarm-browser
```

### Views

1. **Clusters View**: Select a Docker Swarm cluster to connect to
2. **Stacks View**: Browse all stacks in the selected cluster
3. **Services View**: View services within a selected stack
4. **Tasks View**: See all tasks (containers) for a selected service
5. **Container View**: Attach to a running container for interactive shell access

## Development

### Requirements

- Go 1.24.0 or higher
- SSH access to Docker Swarm clusters
- Docker client libraries

### Project Structure

```
swarm-browser/
├── main.go                 # Application entry point
├── internal/
│   ├── app/               # TUI application logic
│   ├── config/            # Configuration management
│   ├── core/              # Core business logic
│   └── services/          # Service layer (Docker, SSH)
├── clusters.yml           # Cluster configuration
└── install.sh            # Installation script
```

### Running in Development

```bash
# Run the application
go run main.go

# Run with development clusters
./run-dev.sh

# Run tests
go test ./...

# Build binary
go build -o swarm-browser main.go
```

### Debugging

Debug logs are written to `debug.log` in the working directory. To monitor logs in real-time:

```bash
tail -f debug.log
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The Elm Architecture for Go
- Docker client library from [moby/moby](https://github.com/moby/moby)

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/Mendes11/swarm-browser/issues) on GitHub.