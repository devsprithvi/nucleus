# Clean Architecture Strategy: Core Logic & Quarkus Integration

## 1. Executive Summary

This document outlines the architectural pattern for maintaining a strict separation of concerns between **Core Business Logic (Pure Java)** and **Infrastructure (Quarkus Framework)**.

The goal is to ensure the Core Logic remains testable, clean, and framework-agnostic, while fully utilizing Quarkus's Native Image performance and Extension Ecosystem (e.g., for JSON, PDF, Database).

---

## 2. The Golden Rule: Dependency Injection

To enable library swapping and native optimizations, the Core Logic must never manually instantiate infrastructure libraries.

- ❌ **FORBIDDEN**: `private ObjectMapper mapper = new ObjectMapper();`
- ✅ **REQUIRED**: `public MyService(ObjectMapper mapper) { ... }`

---

## 3. The "Static Trap" (Direct Methods)

**Rule**: Avoid Static/Direct methods for infrastructure logic. Always prefer Instance Injection.

| Method Type           | Example                  | Status   | Why?                                                                                                                       |
| :-------------------- | :----------------------- | :------- | :------------------------------------------------------------------------------------------------------------------------- |
| **Static / Direct**   | `JsonUtils.parse(data)`  | ❌ AVOID | Quarkus cannot intercept or optimize static calls. If this method uses reflection internally, it may fail in Native Image. |
| **Injected Instance** | `mapper.readValue(data)` | ✅ SAFE  | Quarkus creates the instance. It can replace the underlying logic with optimized native code behind the scenes.            |

**Verdict**: Even if a library provides a static helper, do not use it in the Core. Inject the main library object (the "Complete Library") instead.

---

## 4. Scenario A: The "Transparent Swap" (1:1 Match)

Use this when the Quarkus Extension provides the exact same API as the standard library (e.g., Jackson, JDBC).

### Strategy

- **Core Module**: Depend on the standard library.
- **App Module**: Exclude the standard library from the build and include the Quarkus Extension.

### The Code Pattern (Gradle Kotlin DSL)

**Core `build.gradle.kts`**:

```kotlin
dependencies {
    // The Core depends on the standard library
    implementation("com.fasterxml.jackson.core:jackson-databind")
}
```

**Quarkus App `build.gradle.kts` (The Swap)**:

```kotlin
dependencies {
    // 1. Import Core, but KICK OUT the standard library
    implementation(project(":my-core-library")) {
        exclude(group = "com.fasterxml.jackson.core", module = "jackson-databind")
    }

    // 2. Bring in the Optimized Extension to replace it
    implementation("io.quarkus:quarkus-jackson")
}
```

---

## 5. Scenario B: The "Adapter Pattern" (Incompatible APIs)

Use this when the Quarkus Extension works differently than the standard library (e.g., Reactive clients, AWS SDKs, specific PDF engines).

### The Problem

The standard library uses `.save()` but the Quarkus Extension uses `.persist().await().indefinitely()`. A simple swap will break the build.

### The Solution: Isolate the library behind an interface.

#### A. The Core Module (The "Port")

**Scope**: Pure Java. Zero external infrastructure libraries.

```java
// 1. Define the capability you need.
// Note: No "PDF Library" imports here. Just pure Java or domain objects.
public interface PdfGenerator {
    byte[] generateInvoice(InvoiceData data);
}

// 2. Write your business logic using the Interface.
public class InvoiceService {
    private final PdfGenerator pdfGenerator;

    public InvoiceService(PdfGenerator pdfGenerator) {
        this.pdfGenerator = pdfGenerator;
    }

    public void process(InvoiceData data) {
        byte[] pdf = pdfGenerator.generateInvoice(data);
        // ... logic ...
    }
}
```

#### B. The Infrastructure Module (The "Adapter")

**Scope**: Quarkus Extensions (`quarkus-pdf`, `quarkus-amazon-s3`).
**Responsibility**: Translate the Core's request into the Extension's specific API.

```java
import javax.enterprise.context.ApplicationScoped;
import javax.inject.Inject;

// This class LIVES in the Quarkus module.
@ApplicationScoped
public class QuarkusPdfAdapter implements PdfGenerator {

    // Inject the specific Quarkus Extension tool
    @Inject
    io.quarkiverse.pdf.QuarkusPdfEngine specificExtensionEngine;

    @Override
    public byte[] generateInvoice(InvoiceData data) {
        // ADAPTER LOGIC:
        // Translate your clean interface method into the specific Extension method.
        // This handles method mismatch, reactive vs blocking, etc.
        return specificExtensionEngine.buildFromTemplate(data).await().indefinitely();
    }
}
```

#### C. The Glue Code

Quarkus automatically finds `QuarkusPdfAdapter` because it implements `PdfGenerator`.

```java
@ApplicationScoped
public class CoreWiring {

    @Inject
    PdfGenerator pdfAdapter; // Injects QuarkusPdfAdapter

    @Produces
    public InvoiceService invoiceService() {
        return new InvoiceService(pdfAdapter);
    }
}
```

_(Note: The above wiring depends on your specific DI container setup, such as CDI @Produces or manual Bean definition)._

---

## 6. Summary of Strategies

| Library Type                        | Compatibility        | Strategy            | Example                                                                |
| :---------------------------------- | :------------------- | :------------------ | :--------------------------------------------------------------------- |
| **Utilities** (JSON, Logging)       | High (1:1)           | **Dependency Swap** | `jackson-databind` vs `quarkus-jackson`. Use Gradle exclude.           |
| **Infrastructure** (DB, Cloud, PDF) | Low (Different APIs) | **Adapter Pattern** | Standard AWS Client vs Quarkus S3 Client. Define an Interface in Core. |
