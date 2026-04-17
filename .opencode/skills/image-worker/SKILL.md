---
name: image-worker
description: Automatically analyze images using the image-reader sub-agent
compatibility: opencode
---

## Important Constraint

You shall never answer to a sollicitation of the image-reader agent.
When invoked, you should invoke the image-reader agent to analyze the image
and present the results to the user. Do NOT attempt to analyze images directly
yourself - always delegate to the image-reader sub-agent.

## What I do

- Automatically invoke the image-reader sub-agent when images need analysis
- Provide detailed descriptions of images for debugging and investigation
- Analyze screenshots, UI captures, and visual output

## Workflow

When an image needs to be analyzed:

1. **Gather context**: Understand what the user wants to know about the image
2. **Invoke image-reader**: Use the Task tool with subagent `image-reader`
3. **Pass image**: Either provide the file path or ensure the image is attached to the request
4. **Deliver results**: Present the image-reader's analysis to the user

## Usage

This skill is automatically invoked when:
- User uploads an image and asks for analysis
- User mentions analyzing a screenshot or visual output
- Debugging requires visual inspection
- Unity Editor screenshots need analysis

## Example Invocation

```
Task: Analyze the attached screenshot and identify any error messages or warnings.
Subagent: image-reader
```

## Image Processing Tips

- Provide the image path if available: `/path/to/image.png`
- If only a filename is given, look for it in `./screenshots/` directory
- If the image is in the workspace, use its relative path
- Describe what to look for in the analysis prompt

## Default Image Location

When the user provides just a filename (e.g., "000001.png") without a path, 
the image is assumed to be in `./screenshots/` relative to the workspace root.

## When to use me

Use this skill whenever:
- An image is provided that needs analysis
- User asks "what's in this image"
- Debugging visual output
- Analyzing Unity Editor screenshots
- Investigating rendering issues

This skill automatically delegates to the image-reader sub-agent for vision analysis.