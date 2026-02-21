# Project Execution Strategy & Phasing

**Objective:** Define the engineering approach for handling Technical Debt, Task Management, and Project Phases.
**Scope:** Applies to all contributors during the MLP (Minimum Lovable Product) and Maintenance phases.

---

## 1. Technical Debt Strategy ("The Context Chain")
We prioritize velocity over perfection during the MLP phase, but we never lose context. We use a strict **Dependency Chain** to manage "temporary solutions" (Swimming) vs. "permanent fixes" (Bridges).

### The Protocol:
1.  **Level 1 (Code Anchor):**
    * If a temporary fix is implemented, tag it in the code:
    * `// TODO: Temporary implementation. SEE: .todo/HF_PROVIDER_RESEARCH.md`
2.  **Level 2 (Context Record):**
    * Create a file in `.todo/` that captures the *Mental State* (Research, Dead Ends, Logic).
    * *Rule:* Do not use sparse notes. Write engineering logs.
3.  **Level 3 (System Dependency):**
    * If the fix requires a tool change (e.g., Jira setup), the Context Record must link to the System Spec.
    * *Example:* "Blocked by: `.todo/TOOLING_SETUP.md`"

**Result:** The code remains clean of essay-length comments, but the debt is fully documented and chained in the repo.

---

## 2. Phased Task Management
We adhere to a "Phase-Dependent" strictness model. We do not use Jira/Tickets when they reduce velocity.

### Phase 1: The MLP (Current State)
* **Goal:** Creation & Lovability.
* **Source of Truth:** The TypeSpec (`.tsp`) and Code.
* **Ticket Policy:** **ZERO TICKETS.**
    * Tasks "emerge" naturally. If the code deviates from the TypeSpec, the task is "Fix the Code."
    * *Prohibited:* Creating tickets for "Implement Feature X."
    * *Allowed:* Creating tickets only for external blockers that stop development completely.

### Phase 2: Post-Launch Maintenance
* **Trigger:** V1.0 is live in production with active users.
* **Goal:** Stability & Iterate.
* **Source of Truth:** Issue Tracker (Jira).
* **Ticket Policy:** **100% COVERAGE.**
    * We shift from "Artist" (creation) to "Curator" (maintenance).
    * All bugs, feature requests, and refactors must have a Ticket ID.

---

## 3. The "Natural Hierarchy" Rule
We reject artificial folder structures (Epics/Stories) in favor of **Logical Dependency Chains**.

* **Anti-Pattern:** Nesting Task B inside Task A just to group them.
* **Standard:** Linking Task B as **Blocked By** Task A.
* **Benefit:** This creates a realistic execution graph where priority is dictated by blockers, not by arbitrary categories.