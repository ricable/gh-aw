---
# Amplifier - AI-powered modular development assistant
# Microsoft's research-based AI CLI with extensible modules and specialized agents
#
# Documentation: https://github.com/microsoft/amplifier
# Official Docs: https://microsoft.github.io/amplifier-docs/
#
# This shared workflow provides:
# - Automatic amplifier installation via uv tool
# - Bash tool access for running amplifier commands
# - Configuration for AI provider credentials (Anthropic, OpenAI, Azure OpenAI, Ollama)
#
# Available agents:
#   - zen-architect: System design with ruthless simplicity
#   - bug-hunter: Systematic debugging
#   - web-research: Web research and content fetching
#   - explorer: Breadth-first exploration of local code with summaries
#   - modular-builder: Code implementation
#
# Usage:
#   imports:
#     - shared/mcp/amplifier.md

tools:
  bash:
    - "amplifier *"
    - "uv *"

steps:
  - name: Install UV package manager
    id: setup-uv
    run: |
      if ! command -v uv &> /dev/null; then
        echo "Installing UV package manager..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
        export PATH="$HOME/.cargo/bin:$PATH"
        echo "$HOME/.cargo/bin" >> $GITHUB_PATH
      fi
      uv --version
      echo "UV package manager is ready"

  - name: Install Amplifier
    id: setup-amplifier
    run: |
      echo "Installing Amplifier from GitHub..."
      uv tool install git+https://github.com/microsoft/amplifier
      export PATH="$HOME/.local/bin:$PATH"
      echo "$HOME/.local/bin" >> $GITHUB_PATH
      amplifier --version || echo "Amplifier installed (version check not available)"
      echo "Amplifier is ready"
      mkdir -p /tmp/gh-aw/amplifier

  - name: Configure Amplifier Provider
    id: configure-amplifier
    run: |
      # Check which AI provider credentials are available
      if [ -n "$ANTHROPIC_API_KEY" ]; then
        echo "Anthropic credentials detected"
        export AMPLIFIER_PROVIDER="anthropic"
      elif [ -n "$OPENAI_API_KEY" ]; then
        echo "OpenAI credentials detected"
        export AMPLIFIER_PROVIDER="openai"
      elif [ -n "$AZURE_OPENAI_ENDPOINT" ]; then
        echo "Azure OpenAI credentials detected"
        export AMPLIFIER_PROVIDER="azure-openai"
      else
        echo "No AI provider credentials detected. Amplifier will require configuration."
        echo "To use Amplifier, configure one of the following secrets:"
        echo "  - ANTHROPIC_API_KEY (recommended)"
        echo "  - OPENAI_API_KEY"
        echo "  - AZURE_OPENAI_ENDPOINT with AZURE_OPENAI_API_KEY"
      fi
      
      # Create amplifier config directory if it doesn't exist
      mkdir -p ~/.config/amplifier
      
      echo "Amplifier provider configuration complete"
    env:
      ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
      OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
      AZURE_OPENAI_ENDPOINT: ${{ secrets.AZURE_OPENAI_ENDPOINT }}
      AZURE_OPENAI_API_KEY: ${{ secrets.AZURE_OPENAI_API_KEY }}
---

# Amplifier Usage Guide

Microsoft Amplifier is an AI-powered modular development assistant with specialized agents for different tasks. It has been installed and is available via the `amplifier` command.

A temporary folder `/tmp/gh-aw/amplifier` is available for caching intermediate results.

## Configuration

Amplifier supports multiple AI providers. Configure your preferred provider using repository secrets:

### Anthropic Claude (Recommended)
```yaml
secrets:
  ANTHROPIC_API_KEY: your-api-key
```
Get an API key at: https://console.anthropic.com/settings/keys

### OpenAI
```yaml
secrets:
  OPENAI_API_KEY: your-api-key
```
Get an API key at: https://platform.openai.com/api-keys

### Azure OpenAI (Enterprise)
```yaml
secrets:
  AZURE_OPENAI_ENDPOINT: https://your-resource.openai.azure.com/
  AZURE_OPENAI_API_KEY: your-api-key
  # Or use Azure CLI authentication with managed identity
```

### Ollama (Local, Free)
No API key required. Make sure Ollama is running:
```bash
ollama serve
ollama pull llama3
```

## Common Amplifier Operations

### Initialize Amplifier
```bash
# First-time setup wizard (auto-detects missing config)
amplifier init

# Or configure programmatically
amplifier provider use anthropic --model claude-sonnet-4-5
# or
amplifier provider use openai --model gpt-5.2
```

### Single Command Execution
```bash
# Get quick answers
amplifier run "Explain async/await in Python"

# Generate code
amplifier run "Create a REST API for a todo app with FastAPI"

# Debug issues
amplifier run "Why does this code throw a TypeError: [paste code]"
```

### Using Specialized Agents

Amplifier includes several specialized agents for focused tasks:

```bash
# Use zen-architect for system design
amplifier run "Design a caching layer with careful consideration"

# Use bug-hunter for systematic debugging
amplifier run "Use bug-hunter to debug this error: [paste error]"

# Use explorer for code exploration
amplifier run "Use explorer to analyze the project structure"

# Use web-research for online research
amplifier run "Use web-research to find best practices for error handling in Go"
```

### Working with Bundles

Bundles provide additional capabilities:

```bash
# Add capability bundles
amplifier bundle add git+https://github.com/microsoft/amplifier-bundle-recipes@main
amplifier bundle add git+https://github.com/microsoft/amplifier-bundle-design-intelligence@main

# List available bundles
amplifier bundle list

# Use a specific bundle
amplifier bundle use recipes

# See current bundle
amplifier bundle current
```

### Sessions & Persistence

```bash
# Resume most recent session
amplifier continue

# Resume with new prompt
amplifier continue "follow-up question"

# List recent sessions (current project only)
amplifier session list

# List all sessions across all projects
amplifier session list --all-projects

# View session details
amplifier session show <session-id>

# Resume a specific session
amplifier session resume <session-id>
```

## Available Agents

The **foundation** bundle (default) includes these specialized agents:

- **zen-architect**: System design with ruthless simplicity
- **bug-hunter**: Systematic debugging and error resolution
- **web-research**: Web research and content fetching
- **explorer**: Breadth-first exploration of local code, docs, and files with citation-ready summaries
- **modular-builder**: Code implementation
- **git-ops**: Git operations and version control
- And more...

Use `/agents` in interactive mode to see all available agents.

## Interactive Chat Mode

```bash
# Start a conversation
amplifier

# In chat mode:
# - Context persists across messages
# - Use /help to see available commands
# - Use /tools, /agents, /status, /config to inspect session
# - Use /think and /do to toggle plan mode
# - Type 'exit' or Ctrl+C to quit
```

## Best Practices

1. **Provider Selection**: Anthropic Claude is most tested and recommended
2. **Session Management**: Use session persistence for complex multi-step tasks
3. **Agent Delegation**: Let Amplifier choose the right agent, or specify explicitly when needed
4. **Bundles**: Start with the `foundation` bundle (default) which includes everything for development
5. **Timeouts**: Amplifier operations may take time depending on the AI provider. Ensure adequate workflow timeouts.

## Notes

- Amplifier is a research demonstrator in early preview
- Use with caution and careful human supervision
- Amplifier works best on macOS, Linux, and Windows Subsystem for Linux (WSL)
- Native Windows shells have known issues—use WSL unless actively contributing Windows fixes

## More Information

- GitHub Repository: https://github.com/microsoft/amplifier
- Documentation: https://microsoft.github.io/amplifier-docs/
- Log Viewer (for debugging): https://github.com/microsoft/amplifier-app-log-viewer

<!--
Amplifier MCP Server Integration Note:

While Amplifier is not itself an MCP server, it is designed to be extensible and can work
with MCP servers as a client. This shared workflow configures Amplifier as a CLI tool that
can be used within GitHub Actions workflows.

For MCP server integration, Amplifier can be configured to use MCP tools through its 
modular architecture, but that configuration would need to be done within the Amplifier
session itself, not through this shared workflow file.
-->
