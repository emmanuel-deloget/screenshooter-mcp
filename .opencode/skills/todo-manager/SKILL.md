---
name: todo-manager
description: Manage todo items across internal todolist and TODO-*.md files
compatibility: opencode
---

## What I do

- Keep the internal todolist in sync with TODO-*.md files in the project root
- Add new todo items to both the internal todolist and the appropriate TODO-*.md file
- After compaction or at session start, reconcile the internal todolist with TODO-*.md files

## When I'm Triggered

This skill is automatically invoked when:

- You use the TodoWrite tool to add, update, or complete tasks
- A new session starts (after compaction or at the beginning of a conversation)
- The user explicitly asks to add a new todo item

## Workflow

### Adding a New Todo Item

When adding a new todo item:

1. Use the TodoWrite tool to update the internal todolist
2. Identify which TODO-*.md file the item belongs to (or create a new one)
3. Add the item to the TODO-*.md file following the format:

```
## Summary

Add a line with the format:
- [status] [number]. [Title]

Where:
- status is ✅ (complete), 🔄 (in_progress), or ⏳ (not started)
- number is the next available item number in that file
- Title is a short name for the item
```

4. Add a detailed section at the end:

```
# Item Title

Status: [✅ Complete | 🔄 In Progress | ⏳ Not Started]

[3-5 line explanation of what the item is about and why it matters]
```

### Reconciliation

At the start of a session (or after context compaction):

1. Find all TODO-*.md files in the project root
2. Read each file and extract items from the Summary section
3. Update the internal todolist to match items from TODO files
4. Use the section titles and descriptions to populate the todolist entries

## File Format

Each TODO-*.md file uses this format:

```markdown
# File Title

## Summary

- ✅ 1. First item
- 🔄 2. Second item  
- ⏳ 3. Third item

---

# First Item Title

Status: ✅ Complete

[Explanation - 3-5 lines]

---

# Second Item Title

Status: 🔄 In Progress

[Explanation - 3-5 lines]
```

## Status Characters

- ✅ : Complete / Done
- 🔄 : In Progress / Working on
- ⏳ : Not Started / Pending

## Notes

- The internal todolist uses 'status' field: 'completed', 'in_progress', 'pending'
- TODO-*.md files use emoji: ✅, 🔄, ⏳
- When reconciling, map: completed→✅, in_progress→🔄, pending→⏳
- Each TODO-*.md file maintains its own numbering sequence
- The todo-manager skill should be loaded automatically at session start

## Architecture Update Trigger

When marking a todo item as complete (changing status to ✅):

1. Determine the relevant architecture document:
   - For editor-related items → `Documentation~/editor-architecture.md`
   - For runtime-related items → `Documentation~/runtime-architecture.md`
   - For general/shared items → check both documents

2. Invoke the **architecture-expert** agent to review and update the relevant architecture file:
   - Use the Task tool with `architecture-expert` subagent type
   - Request that they check the architecture document and update it to better describe the implemented feature
   - Focus on documenting new components, interfaces, and design patterns

3. The architecture-expert should update:
   - Class relationships and interfaces
   - Component responsibilities
   - Data flow diagrams
   - File reference appendices