# SSH-Proxy

[![Go Report Card](https://goreportcard.com/badge/github.com/shamir0xe/ssh-proxy)](https://goreportcard.com/report/github.com/shamir0xe/ssh-proxy)
[![GoDoc](https://godoc.org/github.com/shamir0xe/ssh-proxy?status.svg)](https://godoc.org/github.com/shamir0xe/ssh-proxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A robust, self-healing SSH tunnel manager written in Go.

## Overview

`ssh-proxy` is a lightweight utility designed to create and persistently monitor SSH tunnels. It acts as a wrapper around the standard `ssh` command, adding a layer of resilience. The application periodically checks the health of each configured tunnel and, if a connection is found to be down, it automatically attempts to re-establish it.

This is ideal for maintaining stable connections to remote services (like databases or web servers) through a firewall or bastion host.

## Installation

There are a few ways to install `ssh-proxy`.

### 1. Pre-compiled Binaries (Recommended)

The easiest way to get started is to download a pre-compiled binary for your operating system from the [**GitHub Releases**](https://github.com/shamir0xe/ssh-proxy/releases) page.

1.  Go to the latest release.
2.  Download the appropriate archive for your OS and architecture (e.g., `ssh-proxy_1.0.0_linux_amd64.tar.gz`).
3.  Extract the archive.
4.  Move the `ssh-proxy` binary to a directory in your `$PATH`, like `/usr/local/bin`.

### 2. Using `go install`

If you have Go installed, you can build and install the latest version with a single command:
```bash
go install github.com/shamir0xe/ssh-proxy@latest
```
*(Remember to replace `shamir0xe` with your actual GitHub username)*

### 3. Build from Source

You can also build the project from the source code.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/shamir0xe/ssh-proxy.git
    cd ssh-proxy
    ```

2.  **Build the binary:**
    ```bash
    go build -o ssh-proxy .
    ```

3.  **Run it:**
    ```bash
    ./ssh-proxy
    ```

## Configuration

`ssh-proxy` is configured using a `config.yaml` file in the same directory as the executable.

1.  **Create your config file:**
    ```bash
    cp config.sample.yaml config.yaml
    ```
2.  **Edit `config.yaml`** to define your tunnels.

## Usage

Once configured, simply run the executable:

```bash
./ssh-proxy
```

The application will read the `config.yaml`, start all enabled tunnels, and begin monitoring them. It will log its status to standard output. For long-running processes, consider running it with a process manager like `systemd` or `supervisor`.

## Contributing

Contributions are welcome! If you have a feature request, a bug report, or a pull request, please feel free to open an issue or PR on the GitHub repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
