# Tooling Specification: Jira Project for Internal Tools

**Objective:** Setup a tracking system capable of handling "Research-Heavy" dev tasks.
**Current Blocker:** `HF_PROVIDER_RESEARCH.md` cannot be executed without this system.

## 1. Project Configuration
- **Name:** Internal Tools (KEY: TOOL)
- **Type:** Software Development (Kanban)
- **Workflow:** To Do -> Researching -> Coding -> Review -> Done

## 2. Custom Field Requirements
The default Jira fields are insufficient for our research-based workflow. We must add:
1.  **"Dead Ends" (Rich Text):** Mandatory field to list what *didn't* work (stops re-research).
2.  **"Context Link" (URL):** Link to local .todo files or Notion pages.

## 3. Automation Rules
- **Rule:** When status moves to "Researching", automatically assign to me.
- **Rule:** If "Dead Ends" is empty, block transition to "Coding" (Force documentation).

## 4. Migration Plan
Once this project is live:
1. Create Ticket TOOL-1: "Implement Hugging Face Provider".
2. Copy content from `.todo/HF_PROVIDER_RESEARCH.md` into TOOL-1.
3. DELETE the local `.todo` file.