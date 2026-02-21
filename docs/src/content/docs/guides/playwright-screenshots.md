---
title: Taking Website Screenshots with Playwright
description: How to take website screenshots using Playwright, upload them as assets, and post the results in GitHub issues.
sidebar:
  order: 16
---

This guide shows how to take screenshots of a website using Playwright, upload them to your repository as assets, and embed the URLs in a GitHub issue.

## Overview

The workflow follows three steps in order:

1. Take screenshots using the Playwright MCP tool
2. Upload each screenshot using the `upload-asset` safe output
3. Create an issue with the uploaded asset URLs embedded in the body

Asset URLs are only available after uploading, so screenshots must be captured and saved to disk before any upload step runs.

## Workflow Configuration

```aw wrap
---
on:
  workflow_dispatch:
  issues:
    types: [opened]
engine: copilot
permissions:
  contents: read
  issues: write
tools:
  playwright:
    allowed_domains: ["defaults", "example.com"]
  bash:
    - "mkdir*"
safe-outputs:
  upload-asset:
  create-issue:
    title-prefix: "[screenshot] "
    labels: [screenshot, automated]
network:
  allowed:
    - defaults
---

# Take Website Screenshots

Take a screenshot of https://example.com and post the result in a GitHub issue.

## Step 1: Create output directory

Use the bash tool to create a directory for screenshots:

```bash
mkdir -p /tmp/screenshots
```

## Step 2: Take screenshots

Use the Playwright MCP tools to navigate to the site and take a screenshot:

1. Navigate to `https://example.com`
2. Take a full-page screenshot and save it to `/tmp/screenshots/homepage.png`

## Step 3: Upload the screenshot

Use the `upload_asset` tool to upload `/tmp/screenshots/homepage.png`.
Collect the returned URL.

## Step 4: Create issue

Create a GitHub issue titled "Website Screenshot" with the screenshot embedded:

```markdown
### Screenshot

![Homepage screenshot](URL_FROM_UPLOAD_ASSET)
```
```

## Network Configuration

By default, Playwright can only access `localhost` and `127.0.0.1`. To allow access to an external site, add it to `allowed_domains` inside the `playwright:` tool configuration:

```yaml wrap
tools:
  playwright:
    allowed_domains: ["defaults", "example.com", "*.example.com"]
```

The `allowed_domains` list accepts ecosystem bundle names (`defaults`, `github`, `node`, etc.) and individual domain patterns. Subdomains are included automatically.

> [!TIP]
> If you are testing a local server started during the workflow (for example with `npm run preview`), `localhost` is included by default and no additional configuration is required.

For sites outside of known ecosystems, also add the domain to the top-level `network:` block so the agent's own network traffic is allowed:

```yaml wrap
network:
  allowed:
    - defaults
    - "example.com"
```

> [!NOTE]
> The `network:` block controls outbound traffic from the agent process. `playwright.allowed_domains` controls which sites the browser is permitted to visit. Both must allow the domain when accessing an external server.

## Asset Upload

The `upload-asset` safe output uploads files from the workspace or `/tmp` to an orphaned git branch. The tool returns a public `raw.githubusercontent.com` URL you can embed directly in issue bodies or comments.

Declare the safe output in frontmatter:

```yaml wrap
safe-outputs:
  upload-asset:
    allowed-exts: [.png, .jpg, .jpeg]   # default
    max: 10                              # default
```

In the workflow body, instruct the agent to call `upload_asset` with the file path:

```markdown
Upload `/tmp/screenshots/homepage.png` using the `upload_asset` tool and save the returned URL.
```

> [!IMPORTANT]
> Take all screenshots and save them to disk **before** calling `upload_asset`. The URL is only available after the upload completes, and you need it to embed in the issue body.

## Creating the Issue

After collecting the asset URLs, use `create_issue` to post the results:

```yaml wrap
safe-outputs:
  create-issue:
    title-prefix: "[screenshot] "
    labels: [screenshot, automated]
```

In the workflow body, provide the full issue template including the embedded image URL:

```markdown
Create a GitHub issue with:

Title: Website Screenshot - {{ site name }}

Body:
### Screenshot

![Screenshot]({{ URL from upload_asset }})

### Details
- URL: https://example.com
- Captured: {{ current date }}
```

## Complete Example

The following workflow triggers on `workflow_dispatch` and on issue creation, takes a screenshot of a site, and opens a report issue with the image embedded.

```aw wrap
---
on:
  workflow_dispatch:
  issues:
    types: [opened]
engine: copilot
permissions:
  contents: read
  issues: write
tools:
  playwright:
    allowed_domains: ["defaults", "example.com"]
  bash:
    - "mkdir*"
safe-outputs:
  upload-asset:
  create-issue:
    title-prefix: "[screenshot] "
    labels: [screenshot, automated]
network:
  allowed:
    - defaults
    - "example.com"
---

# Website Screenshot Report

Take a screenshot of https://example.com and create a GitHub issue with the results.

## Steps

1. Run `mkdir -p /tmp/screenshots` using the bash tool.

2. Use Playwright to navigate to `https://example.com` and save a full-page
   screenshot to `/tmp/screenshots/homepage.png`.

3. Upload `/tmp/screenshots/homepage.png` using the `upload_asset` tool.
   Save the returned URL as `SCREENSHOT_URL`.

4. Create a GitHub issue using `create_issue` with the following body:

```markdown
### Screenshot

![Homepage](SCREENSHOT_URL)

### Details
- URL: https://example.com
```
```

## Related Documentation

- [Tools](/gh-aw/reference/tools/#playwright-tool-playwright) - Playwright tool configuration
- [Safe Outputs](/gh-aw/reference/safe-outputs/#asset-uploads-upload-asset) - Asset upload reference
- [Network Configuration](/gh-aw/guides/network-configuration/) - Configuring network access
- [Network Access Reference](/gh-aw/reference/network/) - Complete network permissions reference
