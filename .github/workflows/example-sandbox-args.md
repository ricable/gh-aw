---
description: Example workflow demonstrating sandbox.agent.args configuration
on:
  workflow_dispatch:

permissions:
  contents: read

engine: copilot

sandbox:
  agent:
    id: awf
    args: ["--enable-api-proxy"]

timeout-minutes: 5
---

# Sandbox Args Example

This is an example workflow demonstrating the `sandbox.agent.args` configuration.

## Mission

Test that custom AWF arguments are properly passed to the firewall.

## Instructions

1. Check that you are running in a sandboxed environment
2. Confirm network restrictions are in place
3. Report success
