---
on: issues
engine: copilot
network:
  allowed:
    - defaults
  firewall:
    version: v1.0.0
    log-level: debug
    ssl-bump: true
    allow-urls:
      - "https://github.com/githubnext/*"
      - "https://api.github.com/repos/*"
---

# Test Firewall SSL Bump

Test that firewall SSL bump configuration is properly extracted.
