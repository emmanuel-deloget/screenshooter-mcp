---
name: brainstorming
description: Coordinate multiple expert agents for collaborative analysis and brainstorming
compatibility: opencode
---

## What I do

- Coordinate multiple expert agents to provide diverse perspectives
- Facilitate collaborative analysis on architecture and optimization questions
- Synthesize insights from different expertise areas into unified recommendations

## When I'm Triggered

This skill is automatically invoked when you ask questions about:

- **How should I** implement something
- **Best approach** for a task
- **Should I use** a specific pattern or approach
- Questions about **architecture**, **design**, **structure**
- **What's the best way** to organize or structure code
- Questions about **performance**, **optimization**, **efficiency**

## Workflow

When triggered, I:

1. **Analyze the question** to identify if it requires:
   - Architectural perspective (call architecture-expert)
   - Optimization perspective (call optimization-expert)
   - Both perspectives (call both)

2. **Invoke relevant expert(s)** using Task tool:
   - `architecture-expert` for design/structure concerns
   - `optimization-expert` for performance/optimization concerns

3. **Collect perspectives** from each expert

4. **Synthesize** the insights into a coherent response that combines both viewpoints

## Expert Agents

| Expert | Focus Area |
|--------|------------|
| `architecture-expert` | Design patterns, SOLID, modularity, code organization |
| `optimization-expert` | Performance, algorithms, memory, data structures |

## When to Invoke Each Agent

- **architecture-expert**: Questions about code structure, patterns, design principles
- **optimization-expert**: Questions about performance, algorithms, memory efficiency
- **All**: Complex questions involving multiple areas

## Example Usage

**User asks:** "How should I structure this cave generation system?"

**My response:** 
1. Call architecture-expert for modular design perspective
2. Call optimization-expert for performance perspective
3. Combine both into: "For your cave system, I'd recommend X architecture for flexibility, but keep Y optimized for..."

## Response Format

When providing collaborative analysis:

1. **Question Summary**: Brief restatement of the problem
2. **Architecture Perspective**: Key points from architecture-expert
3. **Optimization Perspective**: Key points from optimization-expert
4. **Combined Recommendation**: Synthesized guidance incorporating both

## Key Principles

- Always consider both architecture AND optimization
- Acknowledge trade-offs between design elegance and performance
- Provide actionable recommendations
- Highlight potential conflicts between approaches