---
name: git-worker
description: Handle git operations according to project-specific requirements
compatibility: opencode
---

## What I do

- Execute git commands as requested by the user
- Create commits following strict conventions
- Manage branches and configurations

## Git Requirements

### Commit Rules
- **NEVER auto-commit** - only commit when explicitly asked by the user
- **ALWAYS use `-s` or `--signoff` flag** for DCO (Developer Certificate of Origin)
- **Title format**: `subsystem: change description`
- **Message**: explain WHY the change was made, not HOW
- **Fixes clause**: when fixing a problem, add `Fixes: <commit hash> (commit title)` between title and body

### Amending Commits
- **DO NOT remove sign-off** when amending: always use `-s` flag in `git commit --amend`

### Configuration
- **ALL config changes** MUST use `--local` flag: `git config --local ...`

### Multiple Commits
- When a lot of work is done before committing, multiple commits may be needed
- **SHOW THE PLAN** before proceeding with any commit

## Workflow

1. Before any commit operation:
   - Run `git status` to see untracked/modified files
   - Run `git diff` to see changes
   - Run `git log` to see recent commits and follow the commit message style

2. Analyze changes and draft commit message:
   - Title: `subsystem: change description` (lowercase, concise)
   - Body: explain WHY, not HOW
   - Add Fixes: clause if fixing a previous issue

3. Stage files and create commit:
   - Use `git add <files>` to stage
   - Use `git commit -s -m "title\n\nbody"` to commit with sign-off

4. After commit, verify with `git status` and `git log`

## Example Commit Message

```
docs: add AGENTS.md for agentic coding guidelines

Add documentation for agentic coding agents operating in this
repository, including build/test commands and code style guidelines
for the Underground Unity package targeting Unity 6.x.

Signed-off-by: Emmanuel Deloget <emmanuel@deloget.com>
```

## When to use me

Use this skill whenever you need to perform git operations. Ask clarifying questions if the requirements are unclear.