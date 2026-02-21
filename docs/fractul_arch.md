# Fractal Architecture: A Universal Recursive Pattern

**Version:** 1.1.0
**Status:** Draft / Proposal

## 1. Executive Summary

This document defines **Fractal Architecture**, a recursive software design pattern that prioritizes strict hierarchy, framework agnosticism, and universal structure. Unlike traditional "layered" architectures that flatten complexity, Fractal Architecture embraces the natural depth of a domain. It treats every unit of code—from the root application down to the smallest sub-feature—as a self-similar structure.

This architecture is universal (applicable to Go, TypeScript, Java, etc.) but is detailed here with specific implementation guidelines for the Java/Kotlin ecosystem, leveraging the Multi-Module Monorepo approach.

## 2. Core Concepts & Terminology

To avoid ambiguity between "modules" (Java/Gradle artifact) and "modules" (Logical grouping), we adopt a strict building metaphor:

| Term              | Analogy        | Definition                                                                                                        | Scope                                              |
| :---------------- | :------------- | :---------------------------------------------------------------------------------------------------------------- | :------------------------------------------------- |
| **Project**       | _The Building_ | The entire repository containing all code, configurations, and tools.                                             | The Root Directory (Monorepo)                      |
| **Build Unit**    | _The Floors_   | A physical compilation unit. In Java/Gradle, this is a distinct project with its own `build.gradle.kts`.          | Top-level separation (e.g., Client, Core)          |
| **Component**     | _The Rooms_    | A logical grouping of domain features. It is not a separate build unit but a strict package/folder encapsulation. | Inside `internal/components` (e.g., Auth, Billing) |
| **Sub-Component** | _The Almirah_  | A recursive child component nested within a parent component.                                                     | Inside `billing/components/invoicing`              |

## 3. The Structural Blueprint

### 3.1 The Physical Hierarchy (Gradle / File System)

The architecture dictates a clear separation between the **Runner** (Framework) and the **Logic** (Core).

```text
Project (Root)
├── settings.gradle.kts                 <-- Defines the Build Units
├── client/                             <-- [Build Unit] The Framework Runner
│   ├── build.gradle.kts                <-- Depends on :core
│   └── src/main/kotlin/...             <-- "Glue Code" only. No business logic.
└── core/                               <-- [Build Unit] The Logic Monolith
    ├── build.gradle.kts                <-- Self-contained. No framework dependencies.
    └── src/main/kotlin/
        └── com/org/app/
            ├── sdk/                    <-- User-Facing API Construction
            └── internal/               <-- The Engine (Implementation)
                ├── domain/             <-- Parent Domain Entities
                ├── infra/              <-- Parent Infrastructure Interfaces
                ├── application/        <-- Parent Application Logic
                └── components/         <-- The Components (Recursive Start)
                    ├── auth/
                    │   └── [Standard Layout]
                    └── billing/
                        └── [Standard Layout]
```

**Note on SDK vs. Internal:**

- **`internal/`**: Contains the complete action components and business logic. It is the "Engine."
- **`sdk/`**: Uses the internal components to construct the polished, user-facing API. It is the "Driver's Seat." It is not just interfaces; it is the composition layer that prepares the logic for consumption.

### 3.2 The Universal "Standard Layout"

Every **Component** (whether it is the top-level `internal` block or a deep billing feature) must adhere to the **4-Folder Layout**. This consistency allows developers to navigate any level of depth instantly.

- **`domain/`**
  - **Content:** Pure Entities, Value Objects, Domain Events.
  - **Rule:** Zero dependencies. Pure Kotlin/Java.

- **`infra/`**
  - **Content:** Interface Implementations (Repositories, Gateways), Adapters.
  - **Rule:** Can use standard libraries, but avoids proprietary framework annotations (Quarkus/Spring) in the core logic.

- **`application/`**
  - **Content:** Use Cases, Services, Orchestration logic.
  - **Rule:** Coordinates Domain and Infra.

- **`components/` (Optional)**
  - **Content:** Child Components (Sub-features).
  - **Rule:** The recursive entry point. Previously referred to as "modules", we name this folder `components` to avoid confusion with Build Units.

## 4. Scope & Boundary Rules (The "Fractal Laws")

Since Components exist within a single Build Unit, we enforce boundaries via strict discipline or automated testing.

### Rule A: The Downward Flow (Parent → Child)

- **Definition:** A Parent Component (e.g., Core) can only access the **Public Interface** (SDK/API) of its Child Component (Billing).
- **Restriction:** The Parent must not import classes from the Child's `internal`, `infra`, or `domain` folders directly unless exposed via the Child's SDK.

### Rule B: The Upward Flow (Child → Parent)

- **Definition:** A Child Component (Billing) inherits context from its Parent.
- **Permission:** The Child can access the Parent's `domain` folder.
- **Example:** Billing needs the `User` entity defined in the Parent Core `domain`. This is allowed and encouraged to share common language.

### Rule C: The Sibling Wall (Horizontal)

- **Definition:** Components at the same level (Auth and Billing) are Siblings.
- **Restriction:** Siblings cannot touch each other's internal state.
- **Communication:**
  - Via Parent (Mediator Pattern).
  - Via Public SDK (if strictly defined).
  - **Never** via internal implementation imports.

## 5. Implementation Guide: Framework Agnosticism

We separate the **Core Logic** (Pure) from the **Framework** (Quarkus/Spring) to ensure longevity and portability.

### 5.1 The Core (Pure Logic)

Located in: `:core` Build Unit. No `@ApplicationScoped`, No `@Inject`. Use pure Constructor Injection.

```kotlin
// File: core/src/.../internal/components/billing/application/BillingService.kt
package com.org.app.internal.components.billing.application

import com.org.app.internal.components.billing.domain.Bill
import com.org.app.internal.components.billing.infra.BillingRepository

// PURE KOTLIN. No Framework Imports.
class BillingService(
    private val repository: BillingRepository,
    private val taxCalculator: TaxCalculator
) {
    fun process(amount: Double): Bill {
        val tax = taxCalculator.calculate(amount)
        val bill = Bill(amount, tax)
        return repository.save(bill)
    }
}
```

### 5.2 The Client (The Wiring)

Located in: `:client` Build Unit. This is the only place where the Framework exists. It bridges the gap.

```kotlin
// File: client/src/.../config/BillingConfiguration.kt
package com.org.client.config

import jakarta.enterprise.context.ApplicationScoped
import jakarta.enterprise.inject.Produces
import com.org.app.internal.components.billing.application.BillingService
import com.org.app.internal.components.billing.infra.BillingRepository

@ApplicationScoped
class BillingConfiguration {

    // The Framework creates the Pure Logic instance here
    @Produces
    @ApplicationScoped
    fun provideBillingService(
        repo: BillingRepository,
        calc: TaxCalculator
    ): BillingService {
        return BillingService(repo, calc)
    }
}
```

## 6. Automated Architectural Enforcement

To maintain Semantic Consistency across the project, we use Architecture Testing tools (specifically **ArchUnit** in Java/Kotlin) to enforce the Fractal Laws automatically. This ensures no developer accidentally violates the component boundaries.

```kotlin
// File: core/src/test/.../ArchitectureTest.kt

// Rule: We target the "components" package, not modules
val components = classes()
    .that().resideInAPackage("..components..")

// Rule: Components should be isolated from siblings
val noSiblingAccess = slices()
    .matching("..components.(*)..")
    .should().notDependOnEachOther()

// Rule: Infra should not leak into Domain
val domainIsolation = classes()
    .that().resideInAPackage("..domain..")
    .should().onlyDependOnClassesThat()
    .resideInAnyPackage("..domain..", "java..", "kotlin..")
```

## 7. Universal Application (Polyglot)

The beauty of Fractal Architecture is that the Folder Topology remains identical across languages.

| Language        | Build Unit      | Component      | Wiring Mechanism                      |
|:----------------|:----------------|:---------------|:--------------------------------------|
| **Java/Kotlin** | Gradle Project  | Package        | `@Produces` / Spring `@Bean`          |
| **Go**          | `go.mod` Module | `internal/pkg` | `wire` / `main.go` manual composition |
| **TypeScript**  | NPM Workspace   | Folder         | `inversify` / Factory Functions       |

## 8. Bonus: Dependency Injection & Library Abstraction

A common concern with "Pure Logic" is losing the power of framework-specific libraries (e.g., Quarkus Redis, Quarkus PDF). We solve this using **Interface Abstraction**.

**The Pattern:**

1.  **In Core:** Define an **Interface** for the capability (e.g., `PdfGenerator`).
2.  **In Core Logic:** Use the Interface.
3.  **In Client:** Implement the Interface using the powerful Framework extension (e.g., `QuarkusPdfGenerator` which might use native image optimizations).
4.  **Wiring:** Inject the Framework Implementation into the Core Logic via the Configuration class.

**Benefit:** You get the raw performance and ease of Quarkus Extensions without coupling your Core Domain to Quarkus source code.

## 9. Conclusion

Fractal Architecture provides a disciplined, scalable, and professional mental model for software.

- **Hierarchy:** Solves complexity by nesting it (Rooms in Floors in Buildings).
- **Agnosticism:** Protects the core asset (Logic) from the volatile commodity (Framework).
- **Clarity:** The "4-Folder Layout" creates a navigable map for any developer, instantly.

This document serves as the canonical reference for implementing this pattern.
