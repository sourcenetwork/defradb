---
description: General coding practices and agent interaction rules applicable across the organization, regardless of language.
globs:
alwaysApply: true
---

# Organization General Practices

## I. Core Coding Principles

- **Simplicity:** Prioritize simple, understandable solutions.
- **DRY (Don't Repeat Yourself):** Avoid code duplication. Verify existing functionality before adding new code.
- **Environment Awareness:** Code should correctly handle different environments (dev, test, prod).
- **Cautious Evolution:** When fixing bugs, prefer existing patterns. If new patterns/technologies are necessary, replace and remove old implementations.
- **Concise Files:** Aim for focused files; refactor if they exceed 200-300 lines.
- **Integrated Solutions:** Favor integrated solutions over one-off scripts.

## II. Data & Security

- **No Mock Data (Dev/Prod):** Use mocked/stubbed data _only_ for automated tests.
- **Secure Secrets Management:** Never commit secrets (API keys, passwords) to repositories. Use environment variables or designated secret stores.
- **Security Best Practices:** Adhere to security best practices, especially for user input, auth, and external service interactions. (Consider linking to internal security guidelines if available).

## III. Tooling & Documentation

- **Non-Interactive Execution:** Ensure command-line tools run non-interactively (e.g., use `| cat` or appropriate flags).
- **Thorough Inline Comments:** Document functions, methods, types, classes, and complex logic clearly.
- **Standardized READMEs:** Conform to the [standard-readme](mdc:https:/github.com/RichardLitt/standard-readme) specification for README files.

## IV. Agent Workflow & Task Management

### AI Working Directory Structure

All files related to current work must be managed under the folder `ai/context/<git-branch-name>` (e.g., `ai/context/feat/sync-docs`). This is called the **AI working directory**.

Required files in the AI working directory:

  1. **`specs.md`** - Task specifications
     - Contains all requirements and specifications for the task
     - Aligns with GitHub issues
     - Must include clear acceptance criteria
     - Secure developer approval before implementation

  2. **`plan.md`** - Development plan
     - Describes what will change and where (major architectural decisions)
     - Details major components: classes, structs, interfaces, execution flow
     - Include code snippets only for critical logic clarity
     - For simple tasks: single plan.md file is sufficient
     - For complex tasks: plan.md contains high-level overview, with detailed phase plans in separate files (plan_1.md, plan_2.md, etc.)
     - Use checkboxes sparingly, only for major milestones
     - Focus on implementation approach, not task lists

  3. **`backlog.md`** - Future improvements
     - Contains only realistic improvements that are actually planned
     - Focuses on technical debt and necessary enhancements related to current task
     - Includes refactors identified during development but deferred
     - Keep it practical and actionable, not a wishlist

  4. **`decisions.md`** - Decision journal
     - Starts empty and populated DURING development
     - Records decisions made when initial approach doesn't work
     - Documents failed attempts, why they failed, and chosen alternatives
     - Helps prevent repeating same mistakes
     - Should not duplicate information already in plan.md

  5. **`learnings.md`** - Knowledge capture
     - Contains ONLY new discoveries not documented elsewhere
     - Records unique insights about codebase discovered during task
     - Documents undocumented patterns, conventions, or gotchas
     - Should not repeat information from README files or existing docs
     - Keep it concise and focused on genuinely new knowledge

  6. **`pull_request.md`** - PR description draft
     - Used to draft pull request descriptions
     - Prefer plain text over bullet points

  7. **`ideas.md`** (Optional) - Creative ideas
     - Captures interesting but out-of-scope feature ideas
     - Brainstorming and "nice-to-have" features
     - Not commitments, just possibilities for future consideration

### Task Management

- **Focused Task Execution:**

  - Concentrate solely on the assigned task or bug fix.
  - Document identified but separate issues/refactors in `backlog.md` or as code comments but **do not implement** without explicit instruction.

- **Specifications (`specs.md`):**

  - Create for new features or major changes, aligning with GitHub issues.
  - Secure developer approval before implementation.
  - Must include clear acceptance criteria and requirements.

- **Development Plans (`plan.md`):**

  - Detail implementation steps, interactions with existing code, and testing strategies.
  - Break down complex tasks into measurable phases.
  - Link to relevant documentation/examples.
  - Track status with checkboxes (`- [ ]` / `- [x]`).
  - Document dependencies, prerequisites, and potential blockers.
  - Use `backlog.md` for deferred work.
  - Split complex plans into smaller plans/PRs if needed.

- **Error Resolution:**

  - If errors occur (build, test, etc.), analyze the root cause, attempt a fix, and document both the error and the solution.

- **Pull Requests:**

  - Draft PR descriptions as needed in the AI working directory.

- **Mandatory Quality Checks:** All code changes **must** pass repository-specific quality checks (linting, formatting, tests, build, etc.) before task completion.

- **Contextual Awareness & Resource Utilization:**
  - Actively seek out and utilize available documentation, including project-specific guides, repository structure overviews, and information on core technologies or SDKs relevant to the task.
  - Adhere to any established repository-specific rules, conventions, and architectural guidelines.
