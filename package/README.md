# Pencil CLI

Command-line interface for [Pencil](https://pencil.dev) — create and edit `.pen` design files from the terminal. Run the AI agent with a prompt, call MCP tools directly in an interactive shell, batch-process multiple designs, or export to PNG/JPEG/WEBP/PDF. Built on the same editor engine as the desktop app and IDE extension, with full headless rendering, AI image generation, and stock photo support.

## Installation

```bash
npm install -g @pencil.dev/cli
# or
pnpm add -g @pencil.dev/cli
# or
yarn global add @pencil.dev/cli
```

## Authentication

The CLI requires authentication before running agent operations. There are two methods:

### Interactive Login

```bash
pencil login
```

This starts an interactive session where you choose your login method (email + password or email + OTP code). On success the session token is stored in `~/.pencil/session-cli.json`.

### CLI Key (for CI/CD)

Set the `PENCIL_CLI_KEY` environment variable. CLI keys are scoped to an organization and can be created/revoked in the **Developer Keys** section of your organization settings on the Pencil web app.

```bash
PENCIL_CLI_KEY=pencil_cli_... pencil --out design.pen --prompt "Create a form"
```

The CLI key always takes precedence over a stored session token.

### Checking Status

```bash
pencil status
```

Displays the current authentication method, verifies the session with the backend, and shows account details (email, name, organization for CLI keys).

## Quick Start

```bash
# Log in first
pencil login

# Create a new design from scratch
pencil --out design.pen --prompt "Create a login page with email and password fields"

# Modify an existing design
pencil --in existing.pen --out modified.pen --prompt "Add a blue submit button"

# Start an interactive shell (in headless mode)
pencil interactive -o design.pen

# Start an interactive shell (connect to a running Pencil app)
pencil interactive -a desktop -i design.pen

# List available models
pencil --list-models
```

## Usage

```
pencil [command] [options]

Commands:
  login                  Log in interactively (email + password or OTP)
  status                 Check authentication status
  version                Show CLI version
  interactive            Start an interactive tool shell (see below)

Options:
  --in, -i <path>        Input .pen file (optional, starts with empty canvas if omitted)
  --out, -o <path>       Output .pen file path (required unless --export is used)
  --prompt, -p <text>    Prompt for the AI agent (required)
  --model, -m <id>       Model to use (default: claude-opus-4-6)
  --custom, -c           Use custom Claude model config (e.g. AWS Bedrock, Vertex AI)
  --list-models          List available models and exit
  --tasks, -t <path>     JSON tasks file for batch operations
  --workspace, -w <path> Workspace folder path to run the agent in
  --export, -e <path>    Export an image of the final result
  --export-scale <n>     Export scale factor (default: 1)
  --export-type <type>   Export format: png, jpeg, webp, pdf (default: png)
  --verbose-mcp          Log full MCP tool error details to the console
  --help, -h             Show help message
```

## Interactive Mode

The interactive shell lets you call MCP tools directly on `.pen` files — useful for scripting, debugging, and agentic workflows that need fine-grained control over design operations.

```
pencil interactive [options]

Options:
  --app, -a <name>   Connect to a running Pencil app (e.g. desktop, vscode)
  --in, -i <path>    Input .pen file (optional, empty canvas if omitted)
  --out, -o <path>   Output .pen file (required in headless mode)
  --help, -h         Show detailed tool reference. Important for agentic workflows, so agents can learn the tool.
```

### Modes

**App mode** — connects to a running Pencil desktop or extension. Changes are applied live.

```bash
pencil interactive -a desktop -i my-design.pen
```

**Headless mode** — spins up a local editor without a GUI. Use `save()` to write to `--out`.

```bash
# New empty canvas saved to the output file
pencil interactive -o output.pen

# Edit an existing file
pencil interactive -i input.pen -o output.pen
```

### Shell commands

```
tool_name({ key: value })   Call an MCP tool
save()                      Save the document to disk (headless) or app
exit()                      Exit the shell
```

### Getting started

Begin with `get_editor_state` to load the schema and understand the document:

```
pencil > get_editor_state({ include_schema: true })
pencil > batch_get()                                     # list top-level nodes
pencil > batch_get({ patterns: [{ reusable: true }] })   # find components
```

### Example

```
pencil > get_editor_state({ include_schema: true })
pencil > get_guidelines()
pencil > get_guidelines({ category: "guide", name: "Landing Page" })
pencil > batch_design({ operations: 'hero=I(document,{type:"frame",name:"Hero",x:0,y:0,width:1440,height:900,fill:"#0A0A0A"})' })
pencil > get_screenshot({ nodeId: "hero" })
pencil > save()
pencil > exit()
```

Run `pencil interactive --help` for the full tool reference with parameter types and descriptions.

## Available Models

| Model ID | Description | Default |
|----------|-------------|---------|
| `claude-opus-4-6` | Most capable, higher cost | ✓ |
| `claude-sonnet-4-6` | Fast, balanced performance | |
| `claude-haiku-4-5` | Fastest, lowest cost | |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PENCIL_CLI_KEY` | CLI API key for CI/CD (takes precedence over stored session) |
| `ANTHROPIC_API_KEY` | Your Anthropic API key |
| `PENCIL_API_BASE` | Backend API base URL (default: `https://api.pencil.dev`) |
| `DEBUG` | Enable debug logging |

Further environment variables and agent config can be set in `settings.json`
as well.

## Supported Operations

The CLI supports the following MCP tools with full feature parity to the desktop app:

### Design Operations

| Tool | Description |
|------|-------------|
| `batch_design` | Insert, update, delete, move, copy, replace nodes |
| `batch_get` | Search and read nodes by pattern or ID |
| `get_variables` | Read design variables |
| `set_variables` | Update design variables |
| `get_editor_state` | Get document metadata and structure |
| `snapshot_layout` | Get document structure with computed bounds |
| `find_empty_space_on_canvas` | Find available space for new elements |
| `search_all_unique_properties` | Recursively search for all unique properties on a node tree |
| `replace_all_matching_properties` | Recursively replace all matching properties on a node tree |

### Visual Operations (headless rendering via CanvasKit)

| Tool | Description |
|------|-------------|
| `get_screenshot` | Render a node to PNG image |
| `export_nodes` | Export nodes to images in PNG/JPEG/WEBP/PDF formats |

### Image Generation

The `batch_design` `G()` operation supports both AI-generated and stock images. Generated images are saved to an `images/` directory alongside the output `.pen` file.

| Type | Description |
|------|-------------|
| `G(nodeId, "ai", prompt)` | AI-generated image from a text prompt |
| `G(nodeId, "stock", keywords)` | Stock photo from Unsplash |

### Guidelines

| Tool | Description |
|------|-------------|
| `get_guidelines` | Load guides and styles for working with .pen files |

## Examples

### Create a Login Page

```bash
pencil --out login.pen --prompt "Create a modern login page with:
- Email input field
- Password input field
- 'Sign In' button
- 'Forgot password?' link
- Social login options (Google, GitHub)"
```

### Add Components to Existing Design

```bash
pencil --in dashboard.pen --out dashboard-v2.pen --prompt "Add a sidebar navigation with:
- Dashboard link (active)
- Users link
- Settings link
- Logout button at bottom"
```

### Create a Component Library

```bash
pencil --out components.pen --prompt "Create a component library with:
- Primary, secondary, and ghost button variants
- Text input with label and error state
- Card component with header and content
- Badge component in success, warning, error colors"
```

### Using a Specific Model

Use `--model` to select a different Claude model:

```bash
# Use Claude Opus for complex tasks requiring highest capability
pencil --out complex-app.pen \
  --model claude-opus-4-6 \
  --prompt "Create a complete e-commerce product page with image gallery,
reviews section, related products, and add-to-cart functionality"

# Use Claude Haiku for simple, fast tasks
pencil --out simple.pen \
  --model claude-haiku-4-5 \
  --prompt "Create a simple 404 error page"
```

### CI/CD Usage

```bash
# Authenticate with a CLI key
export PENCIL_CLI_KEY=pencil_cli_...
export ANTHROPIC_API_KEY=sk-ant-...

# Generate designs in a pipeline
pencil --out onboarding.pen --prompt "Create a 3-step onboarding flow"
```

### Verbose MCP Error Logging

Use `--verbose-mcp` to show full MCP tool error details (including stack traces where available) in the CLI output when a tool fails:

```bash
pencil --out debug.pen \
  --prompt "Create a simple layout" \
  --verbose-mcp
```

### Batch Processing with Tasks File

Use `--tasks` to process multiple designs from a JSON tasks file:

```bash
pencil --tasks batch-tasks.json
```

Example `batch-tasks.json`:

```json
{
  "tasks": [
    {
      "out": "landing-page.pen",
      "prompt": "Create a SaaS landing page with hero, features, and pricing sections"
    },
    {
      "in": "existing-app.pen",
      "out": "existing-app-v2.pen",
      "prompt": "Add a dark mode toggle to the header"
    },
    {
      "out": "mobile-menu.pen",
      "model": "claude-haiku-4-5",
      "prompt": "Create a mobile hamburger menu component"
    }
  ]
}
```

Each task in the array supports the same options as CLI arguments:
- `in` - Input file (optional)
- `out` - Output file (required)
- `prompt` - AI prompt (required)
- `model` - Model override (optional)

## Limitations

The CLI is designed for headless operation and has some limitations compared to the desktop app:

- **No real-time preview** - Changes are saved to file, not displayed interactively
- **No interactive UI features** - No selection, zoom, or pan controls
- **No library browsing** - Cannot browse or import from `.pen` libraries

## Token Storage

| File | Purpose |
|------|---------|
| `~/.pencil/session-cli.json` | Stored session token from `pencil login` |

The CLI uses a separate session file from the desktop app (`~/.pencil/session-desktop.json`) so the backend can distinguish which client is in use.

## License

Proprietary — see LICENSE file. © High Agency, Inc.
