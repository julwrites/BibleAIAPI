---
date: 2025-12-11
title: "Agent Harness Adoption"
tags: ["architecture", "harness", "migration"]
created: 2025-12-11 06:20:54
---

Adopted the agent-harness structure. Key changes:
- Added scripts/tasks.py, scripts/memory.py, scripts/bootstrap.py.
- Migrated legacy task files in docs/tasks/ to YAML Frontmatter format.
- Moved docs/tasks/instruction_driven_workflows.md to docs/architecture/ to resolve validation errors.
- Updated AGENTS.md to include harness workflow instructions while preserving project-specific context.
- Added .cursorrules and CLAUDE.md symlink.
- Initialized docs/memories/ and docs/tasks/ directories.
