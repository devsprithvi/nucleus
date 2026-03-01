# Omniscient AI Architecture (Rough Idea)

> **Note:** This document outlines the high-level visionary plan for our AI-assisted development workflow. It serves as a philosophical and architectural compass. The exact algorithms, models, and tooling implementations will likely evolve over time, but the core high-level strategy remains exactly this.

## 1. The Core Philosophy

The shift from "AI as a text auto-completer" to "AI as the Omniscient Architect".

- **The Component Hierarchy:** All development starts by defining a language-agnostic component tree that represents the logical architecture.
- **Library-Centric API Boundaries:** Instead of drafting hypothetical, easily broken abstractions, our Interface Definition Language (IDL—currently TypeSpec, later custom) interfaces must map directly to the advanced native capabilities of the best possible open-source libraries available for the task.

## 2. The "Skeleton" (Omniscient Knowledge Database)

To truly utilize the power of libraries, we cannot rely on human research. Human research relies on "luck", GitHub SEO, and search engines. It leads to using only basic features and constantly reinventing the wheel.

**The Goal:** Build an AI-managed internal "skeleton" database of all worthwhile open-source libraries, so the system inherently knows exactly what API structures to use to implement the component hierarchy perfectly and seamlessly generate code to fill in the blanks.

### The Extraction Process

1. **Targeting Package Registries:** We bypass the noise of the entire broader GitHub ecosystem (empty repos, student exercises) and target curated package managers (e.g., `crates.io`, `npm`, `pkg.go.dev`).
2. **AST & ML Extraction:**
   - The worker downloads a shallow copy of the repository.
   - An Abstract Syntax Tree (AST) parser maps out the exact, deterministic structure of the code.
   - Custom ML Models extract the API boundaries, understanding exactly what the library is capable of—leaving zero room for LLM hallucinations.
   - Upon completion, the raw download is immediately deleted to free up disk space.

## 3. The Most Efficient Execution Pipeline (The Director's Cut)

Mapping millions of repositories can be catastrophically expensive if done with standard generic cloud deployments. We achieve this within a scavenged budget via total resource utilization.

### The Massive Parallel Sprint

We do not run this crawler continuously for years. We run highly concurrent, massive parallel **Sprints** (e.g., executing all processing over 7 days), and schedule these sprints periodically to update our map.

### The Bare Metal Secret

To achieve 100% CPU utilization efficiently:

- Do not use expensive abstracted cloud services (Lambda or managed K8s) that waste resources.
- Use **Raw Bare-Metal Servers** (like Hetzner) or highly discounted Cloud Spot/Preemptible VMs.
- Write the worker engine in a highly concurrent language (Rust or Go). The CPU must scream at 99% utilization—if it is not parsing an AST, it is downloading the next batch. The CPU never sleeps.

### The $0 Ongoing Database Cost

Never pay thousands of dollars a month to host a massive ~5TB Knowledge Graph database in the cloud.

- As the workers run the sprint, they write the extracted structural data (AST representations/mappings) to flat, highly compressed binary chunks (like Parquet or SQLite) inside an ultra-cheap object storage bucket.
- Once the sprint finishes, all cloud servers are destroyed, stopping billing entirely.
- The 5TB Parquet dataset is downloaded physically to a local, $100 NVMe Hard Drive.
- The local AI Architect engine queries the physical drive directly (e.g., using DuckDB) resulting in **$0.00** monthly operational costs for the database.
