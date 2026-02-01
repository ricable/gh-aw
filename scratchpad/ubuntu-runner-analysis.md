# Ubuntu Actions Runner Image Analysis

**Last Updated**: 2026-02-01  
**Source**: [Ubuntu 24.04 Runner Image Documentation](https://github.com/actions/runner-images/blob/ubuntu24/20260126.10/images/ubuntu/Ubuntu2404-Readme.md)  
**Ubuntu Version**: 24.04.3 LTS  
**Image Version**: 20260126.10.1  
**Runner Version**: 2.331.0

## Overview

This document provides an analysis of the default GitHub Actions Ubuntu runner image and guidance for creating Docker images that mimic its environment. The runner image is based on Ubuntu 24.04 LTS and includes a comprehensive suite of development tools, language runtimes, databases, and CI/CD utilities.

## Included Software Summary

The Ubuntu 24.04 runner image includes:
- **Operating System**: Ubuntu 24.04.3 LTS (Noble Numbat)
- **Language Runtimes**: Node.js, Python, Ruby, Go, Java, PHP, Rust, .NET
- **Container Tools**: Docker, containerd, Podman, Kubernetes tools
- **Build Tools**: gcc, g++, clang, CMake, Make, Ninja
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis, SQLite
- **CI/CD Tools**: GitHub CLI, Azure CLI, AWS CLI, Google Cloud SDK, Terraform
- **Testing Tools**: Selenium, Chrome/Firefox drivers, Playwright dependencies

## Operating System

- **Distribution**: Ubuntu 24.04.3 LTS (Noble Numbat)
- **Kernel**: Linux 6.8+ (specific version varies)
- **Architecture**: x86_64
- **Runner Version**: 2.331.0
- **Image Provisioner**: Hosted Compute Agent 20260123.484

## Language Runtimes

### Node.js
- **Versions**: Multiple LTS and current versions typically installed via nvm
  - Common versions: 18.x, 20.x, 22.x (LTS)
  - Latest current release
- **Default Version**: Usually latest LTS (20.x or newer)
- **Package Managers**:
  - npm (bundled with Node.js)
  - yarn 1.x (classic) and 4.x (berry)
  - pnpm (latest)

### Python
- **Versions**: Python 3.x series
  - Python 3.10, 3.11, 3.12 (system default on 24.04)
  - Python 2.7 (legacy, may not be included)
- **Default Version**: Python 3.12
- **Package Manager**: pip (latest)
- **Additional Tools**:
  - pipenv
  - poetry
  - virtualenv
  - pipx

### Ruby
- **Versions**: Multiple versions installed via rvm or rbenv
  - Common versions: 3.0, 3.1, 3.2, 3.3
- **Default Version**: Latest stable (3.3.x)
- **Package Manager**: gem, bundler

### Go
- **Version**: Latest stable release (1.25.x as of 2026-01-26)
- **Package Manager**: go mod (built-in)

### Java
- **Versions**: Multiple JDK versions
  - Temurin (Eclipse Adoptium): 8, 11, 17, 21
  - Default: Java 21 (LTS)
- **Build Tools**:
  - Maven 3.x
  - Gradle 8.x
  - Ant 1.10.x

### PHP
- **Versions**: PHP 8.x series
  - Common versions: 8.1, 8.2, 8.3
- **Default Version**: PHP 8.3
- **Package Manager**: Composer

### .NET
- **SDK Versions**: .NET 6.0, 7.0, 8.0
- **Runtime Versions**: Corresponding runtimes for each SDK

### Rust
- **Version**: Latest stable release via rustup
- **Package Manager**: cargo (bundled)

### Other Languages
- **Perl**: 5.x (system default)
- **R**: Latest release
- **Kotlin**: Latest version
- **Swift**: Latest release for Linux

## Container Tools

### Docker
- **Version**: Docker CE 20.x or newer (typically 27.x)
- **Components**:
  - docker-compose v2.x (Docker Compose CLI plugin)
  - docker-buildx (multi-platform builds)
- **containerd**: Latest stable (1.7.x or newer)
- **Docker Credential Helpers**: Available

### Podman
- **Version**: 4.x or 5.x (rootless container engine)

### Kubernetes Tools
- **kubectl**: Latest stable release
- **helm**: v3.x (latest)
- **minikube**: Latest release
- **kind**: Kubernetes IN Docker - latest
- **k3d**: Lightweight Kubernetes - latest

## Build Tools

- **Make**: 4.3 (system default)
- **CMake**: 3.28 or newer
- **gcc/g++**: 13.x (default on Ubuntu 24.04)
- **clang**: 18.x or newer
- **Ninja**: Latest stable build tool
- **Bazel**: Latest release
- **Autotools**: autoconf, automake, libtool

## Databases & Services

### PostgreSQL
- **Version**: PostgreSQL 16.x (latest stable)
- **Service Status**: Available but not running by default
- **Client Tools**: psql, pg_dump, pg_restore

### MySQL
- **Version**: MySQL 8.4.x (latest)
- **Service Status**: Available but not running by default
- **Client Tools**: mysql, mysqldump

### MongoDB
- **Version**: 8.0.x (latest)
- **Service Status**: Available but not running by default

### Redis
- **Version**: 7.x (latest stable)
- **Service Status**: Available but not running by default

### SQLite
- **Version**: 3.45.x or newer

## CI/CD Tools

### GitHub Tools
- **GitHub CLI (gh)**: Latest release (2.x)
- **GitHub Actions Runner**: 2.331.0

### Cloud Provider CLIs
- **Azure CLI**: Latest (2.x)
- **AWS CLI**: v2 (latest)
- **Google Cloud SDK**: Latest (gcloud, gsutil, bq)

### Infrastructure as Code
- **Terraform**: Latest stable (1.x)
- **Ansible**: Latest via pip
- **Packer**: Latest release
- **Pulumi**: Latest release

### Other DevOps Tools
- **Ansible**: 2.x (latest)
- **Chef**: Latest
- **Puppet**: Latest
- **GitVersion**: For semantic versioning

## Testing Tools

### Browser Testing
- **Selenium Server**: Latest standalone server
- **Google Chrome**: Latest stable
- **Mozilla Firefox**: Latest stable
- **ChromeDriver**: Matching Chrome version
- **GeckoDriver**: Matching Firefox version

### End-to-End Testing
- **Playwright**: Dependencies installed (browsers installed on-demand)
- **Cypress**: Can be installed via npm

### Performance Testing
- **Apache Benchmark (ab)**: System default
- **wrk**: HTTP benchmarking tool

## Version Control

- **Git**: 2.52.0 (from official PPA, latest stable)
- **Git LFS**: Latest release
- **Mercurial**: Latest
- **Subversion**: 1.14.x

## Package Managers

- **apt**: System package manager (dpkg-based)
- **snap**: Universal Linux package manager
- **Homebrew**: Available (linuxbrew)

## Environment Variables

Key environment variables set in the runner:

``````bash
# System paths
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin

# GitHub Actions context
GITHUB_WORKSPACE=/home/runner/work/<repo>/<repo>
RUNNER_TEMP=/home/runner/work/_temp
RUNNER_TOOL_CACHE=/opt/hostedtoolcache

# Go environment
GOROOT=/opt/hostedtoolcache/go/1.25.0/x64
GOPATH=/home/runner/go

# .NET environment
DOTNET_ROOT=/usr/share/dotnet

# Java environment  
JAVA_HOME=/usr/lib/jvm/temurin-21-jdk-amd64

# Node.js environment (via nvm)
NVM_DIR=/home/runner/.nvm

# Container tools
DOCKER_CONFIG=/home/runner/.docker
``````

## Creating a Docker Image Mimic

To create a Docker image that mimics the GitHub Actions Ubuntu runner environment, follow these guidelines. Note that perfectly replicating the entire image is impractical (it's several GB), so focus on the tools your workflows actually need.

### Base Image

Start with the Ubuntu base image matching the runner version:

``````dockerfile
FROM ubuntu:24.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Set timezone (runner uses UTC)
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
``````

### System Setup

``````dockerfile
# Update system packages and install build essentials
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
    build-essential \
    cmake \
    git \
    curl \
    wget \
    unzip \
    software-properties-common \
    ca-certificates \
    gnupg \
    lsb-release \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
``````

### Language Runtimes

``````dockerfile
# Install Node.js (using NodeSource)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

# Install npm global tools
RUN npm install -g yarn pnpm

# Install Python
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Python tools
RUN pip3 install --no-cache-dir pipenv poetry virtualenv

# Install Go (download and extract)
ENV GO_VERSION=1.25.0
RUN curl -LO https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Install Java (Temurin)
RUN wget -O - https://packages.adoptium.net/artifactory/api/gpg/key/public | apt-key add - && \
    echo "deb https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | tee /etc/apt/sources.list.d/adoptium.list && \
    apt-get update && \
    apt-get install -y temurin-21-jdk && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
ENV JAVA_HOME=/usr/lib/jvm/temurin-21-jdk-amd64
ENV PATH=$PATH:$JAVA_HOME/bin

# Install Ruby (using system package or rbenv)
RUN apt-get update && apt-get install -y \
    ruby-full \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
``````

### Container Tools

``````dockerfile
# Install Docker
RUN curl -fsSL https://get.docker.com | sh

# Install Docker Compose
RUN curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# Install Helm
RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
``````

### Additional Tools

``````dockerfile
# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
    chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
    tee /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y gh && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Install AWS CLI
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf aws awscliv2.zip

# Install Terraform
RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
    tee /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && apt-get install -y terraform && \
    apt-get clean && rm -rf /var/lib/apt/lists/*
``````

### Build Tools

``````dockerfile
# Install Maven
RUN apt-get update && apt-get install -y maven && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Gradle
RUN wget https://services.gradle.org/distributions/gradle-8.5-bin.zip -P /tmp && \
    unzip -d /opt/gradle /tmp/gradle-*.zip && \
    rm /tmp/gradle-*.zip
ENV PATH=$PATH:/opt/gradle/gradle-8.5/bin
``````

### Environment Configuration

``````dockerfile
# Set environment variables to match runner
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Create runner-like directory structure
RUN mkdir -p /home/runner/work

# Set working directory
WORKDIR /home/runner/work
``````

### Complete Dockerfile Example

Here's a streamlined Dockerfile focusing on the most common tools:

``````dockerfile
FROM ubuntu:24.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC

# Set timezone
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install system essentials
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
    build-essential \
    cmake \
    git \
    curl \
    wget \
    unzip \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Node.js 20.x LTS
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g yarn pnpm && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Python 3 and tools
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    && pip3 install --no-cache-dir pipenv poetry virtualenv && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GO_VERSION=1.25.0
RUN curl -LO https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Install Docker
RUN curl -fsSL https://get.docker.com | sh && \
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
    chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
    tee /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && apt-get install -y gh && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install kubectl and Helm
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl && \
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Create runner-like directory structure
RUN mkdir -p /home/runner/work

# Set working directory
WORKDIR /home/runner/work

# Keep container running (for testing)
CMD ["/bin/bash"]
``````

## Key Differences from Runner

Important aspects that cannot be perfectly replicated:

1. **GitHub Actions Context**: The runner includes GitHub Actions-specific environment variables (`GITHUB_*` variables) and context that won't be available in a custom Docker image unless you're running inside GitHub Actions.

2. **Pre-cached Dependencies**: The runner image has pre-cached dependencies for faster builds (npm packages, pip packages, Go modules, etc.). Your custom image will need to download these on first use.

3. **Service Configuration**: Some services (PostgreSQL, MySQL, Redis) are pre-installed but not running by default on the runner. They need to be started explicitly in workflows.

4. **File System Layout**: The runner uses specific directory structures:
   - `/home/runner/work/<repo>/<repo>` - Workspace
   - `/opt/hostedtoolcache` - Cached tools and runtimes
   - `/home/runner/work/_temp` - Temporary files

5. **Hosted Tool Cache**: The runner uses `/opt/hostedtoolcache` for versioned tools (Go, Node.js, Python, Ruby). Custom images typically install to standard locations.

6. **Size**: The full runner image is very large (40GB+). Custom images should only include necessary tools to stay lean.

7. **Automatic Updates**: GitHub updates the runner image regularly. Custom images need manual maintenance.

8. **Security**: The runner includes security hardening and monitoring that custom images may lack.

## Usage Recommendations

### When to Use the Official Runner
- Standard CI/CD workflows
- Public repositories
- Workflows that benefit from cached dependencies
- When you need the latest tool versions automatically

### When to Create a Custom Docker Image
- Specific tool versions required
- Custom/proprietary tools needed
- On-premises or self-hosted runners
- Reproducible environments for local development
- When you need minimal images for faster startup

### Best Practices

1. **Start Small**: Only include tools you actually use
2. **Multi-stage Builds**: Use Docker multi-stage builds to reduce final image size
3. **Layer Caching**: Order Dockerfile commands from least to most frequently changed
4. **Version Pinning**: Pin specific versions for reproducibility
5. **Regular Updates**: Update base image and tools regularly for security
6. **Documentation**: Document which tools and versions you've included

## Maintenance Notes

- The runner image is updated regularly by GitHub (usually weekly)
- Check the [actions/runner-images](https://github.com/actions/runner-images) repository for updates
- This analysis should be refreshed periodically to stay current
- Runner image documentation URL format: `https://github.com/actions/runner-images/blob/ubuntu24/<version>/images/ubuntu/Ubuntu2404-Readme.md`
- Image Release Notes: `https://github.com/actions/runner-images/releases/tag/ubuntu24%2F<version>`

## References

- **Runner Image Repository**: https://github.com/actions/runner-images
- **Documentation Source**: https://github.com/actions/runner-images/blob/ubuntu24/20260126.10/images/ubuntu/Ubuntu2404-Readme.md
- **Image Release**: https://github.com/actions/runner-images/releases/tag/ubuntu24%2F20260126.10
- **Ubuntu Documentation**: https://ubuntu.com/server/docs
- **Docker Documentation**: https://docs.docker.com/
- **GitHub Actions Documentation**: https://docs.github.com/en/actions

---

*This document is automatically generated by the Ubuntu Actions Image Analyzer workflow.*
*Last analyzed: 2026-02-01 from workflow run 21565035165*
