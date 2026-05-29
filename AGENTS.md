# AI Agent Operating Guide

## 🔴 Mandatory Read Order (STRICT)

1. README.md
2. context.md
3. project_status.md
4. roadmap.md

---

## 🧠 Source of Truth Rules

Priority:

1. Code
2. project_status.md
3. context.md
4. roadmap.md

- NEVER assume roadmap items are implemented
- If conflict → trust code

---

## 📂 File Responsibilities

### context.md

System architecture and understanding

### roadmap.md

Planned work

### project_status.md

Actual implementation state

---

## ⚙️ Execution Rules

### Before starting:

- Match task with roadmap
- Verify in project_status.md

### After completing:

- Update project_status.md
- Update context.md (if needed)

---

## 🔐 File Access & Modification Rules

### 📖 Read Access

- Full access to all files (code, DB, configs)

### ✏️ Write Access

❗ Requires user approval

### 🛑 Approval Flow

Before changes:

- List files
- Explain changes
- Ask: "Proceed?"

### ✅ After Approval

- Apply changes
- Update docs

### 🚫 Without Approval

- Only suggest changes
- Do not modify files

### ⚠️ DB Changes

Must explicitly warn before execution

---

## ✅ Task Checklist

- Code correct
- Matches roadmap
- Docs updated

## 🧠 Development Rules

### Scope Control

- Implement ONLY the requested roadmap task
- Do not modify unrelated modules
- Do not perform large refactors unless explicitly requested

---

### Architecture Rules

- Preserve existing CLI behavior
- Avoid breaking changes
- Maintain backward compatibility

---

### Simplicity Rules

- Prefer simple implementations
- Avoid overengineering
- Avoid unnecessary abstractions, wrappers, interfaces, or patterns
- Keep code readable and maintainable

---

### Ownership Rule

- The agent is an implementation assistant, not the architecture owner
- Do not introduce new architectural patterns unless explicitly requested

---

### Change Flow

Before changes:

- List affected files
- Explain why each file changes
- Explain expected outcome

---

### Implementation Rules

- Keep changes incremental
- Prefer extraction over rewriting
- Do not move or rename large structures unless required

---

### Testing Rules

After implementation:

- Run relevant tests
- Show test results
- Explain what was validated

---

### Documentation Rules

After task completion:

- Update project_status.md
- Update roadmap.md task status
- Update context.md only if architecture changed

---

### Code Quality

- Add comments for important logic and architectural decisions
- Keep functions focused and small
- Avoid deeply nested logic where possible
