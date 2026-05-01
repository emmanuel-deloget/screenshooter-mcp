# ScreenshooterMCP Agent Skill

This skill enables AI agents to capture and analyze Linux desktop screens (X11 and Wayland).

## Security-related obligation

As an agent, you shall make sure that you get the user authorization before doing any sensitive operation. You MUST ask the user if he agrees to proceed unless he/she instruct you to not ask him for a permission.

Sensitive operations are:
* `capture_screen`, `capture_window`, `capture_region`: any operation that perform a screen, window or region capture may contain confidential or personal information. 
* `analyze_image`, `extract_text`: the image parameter might contain some confidential or personal information.

The `execute_capture_pipeline` tool might execute series of sensitive operation. You don't have to ask the user for each step of the pipeline, but you MUST present hime/her with a clear view of what the pipeline instruction will do and ask him/her for permission before executing the tool (unless he/she instricts you to not ask for a permission).

## Tool Catalog

### Screen Capture

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `list_monitors` | List available displays | none | JSON array of monitors |
| `list_windows` | List open windows with state | none | JSON array of windows |
| `capture_screen` | Capture full screen or monitor | `monitor` (optional) | PNG image |
| `capture_window` | Capture window by title | `title` (required) | PNG image |
| `capture_region` | Capture rectangular region | `x`, `y`, `width`, `height` | PNG image |

### AI Vision Analysis

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `list_vision_providers` | List configured vision providers | none | JSON array of providers |
| `analyze_image` | Analyze image with custom prompt | `image_base64`, `prompt`, `provider` (opt), `timeout` (opt) | Text |
| `extract_text` | Extract text from image (OCR) | `image_base64`, `provider` (opt), `timeout` (opt) | Markdown text |
| `find_region` | Find element coordinates | `image_base64`, `description`, `provider` (opt), `timeout` (opt) | JSON `{x, y, width, height}` |
| `compare_images` | Compare two images | `image_base64`, `image2_base64`, `prompt` (opt), `provider` (opt), `timeout` (opt) | Text |

### Pipeline Execution

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `execute_capture_pipeline` | Chain multiple operations | `pipeline` (array of steps) | Image (base64) and/or Text |

## Common Workflows For Pipeline Execution

### Read text from a specific UI element

```json
{
  "pipeline": [
    {"tool": "capture_window", "parameters": {"title": "Application"}},
    {"tool": "find_region", "parameters": {"description": "the error message"}},
    {"tool": "capture_region", "parameters": {}},
    {"tool": "extract_text", "parameters": {}}
  ]
}
```

### Detect changes after an action

```json
{
  "pipeline": [
    {"tool": "capture_screen", "parameters": {}},
    {"tool": "wait_for", "parameters": {"seconds": 5}},
    {"tool": "capture_screen", "parameters": {}},
    {"tool": "compare_images", "parameters": {}}
  ]
}
```

### Analyze a specific region of the screen

```json
{
  "pipeline": [
    {"tool": "capture_screen", "parameters": {}},
    {"tool": "find_region", "parameters": {"description": "the notification panel"}},
    {"tool": "capture_region", "parameters": {}},
    {"tool": "analyze_image", "parameters": {"prompt": "What does this notification say?"}}
  ]
}
```

### Extract text from a window

```json
{
  "pipeline": [
    {"tool": "capture_window", "parameters": {"title": "Terminal"}},
    {"tool": "extract_text", "parameters": {}}
  ]
}
```

## Pipeline DSL

Each pipeline is an array of steps. Each step has:
- `tool`: The tool to execute
- `parameters`: Tool-specific parameters (optional for tools that use stack input)

### Stack Behavior

| Tool | Pops from stack | Pushes to stack |
|------|-----------------|-----------------|
| `capture_screen` | - | image |
| `capture_window` | - | image |
| `capture_region` | region (if no explicit coords) | image |
| `find_region` | 1 image | text (JSON coords) |
| `extract_text` | 1 image | text |
| `analyze_image` | 1 image | text |
| `compare_images` | 2 images | text |
| `wait_for` | - | - |

- `capture_region` with no `x`/`y`/`width`/`height` parameters pops a region from the stack (output of `find_region`).
- At pipeline end, only the top stack item is returned. Unused items are discarded.
- `wait_for` pauses execution (max 30 seconds), produces no output.

## Vision Provider Selection

- The first provider in the config is the default.
- Specify `provider` to use a different one.
- Use `list_vision_providers` to see available providers.
- Small local models (llava, moondream) may struggle with `find_region`. Use larger models for coordinate tasks.

## Security Notes

- This MCP is **read-only**: it captures screens and analyzes images.
- It does **not** inject keyboard/mouse input.
- It does **not** write files to the filesystem.
- It does **not** modify window state or system configuration (with a small exception: on some systems, it might activate a window before capturing it).
