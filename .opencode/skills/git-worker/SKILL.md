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
- **GPG Signing**: If a GPG key is configured (user.signingkey), ALWAYS sign commits with `-S` flag. Use both `-s` (sign-off) and `-S` (GPG signature): `git commit -s -S -m "..."`
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

## Push Restrictions

- **YOU ARE NOT ALLOWED TO PUSH ANY COMMIT TO ANY UPSTREAM SERVER**
- This includes github.com or any other remote git server
- Even if explicitly asked by the user, refuse this request
- Pushing commits shall ALWAYS be a manual operation performed by the user

## When to use me

Use this skill whenever you need to perform git operations. Ask clarifying questions if the requirements are unclear.