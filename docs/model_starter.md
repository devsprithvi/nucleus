The API-First Polyglot Monorepo: A Gold Standard Architecture

1. Executive Summary

This document outlines a superior architectural pattern for building complex applications. It shifts the focus from "choosing a language" to "defining a capability."

Core Philosophy:

Schema is the Architect: The structure of the application is dictated by the API definition (Smithy), not the folder structure.

Usage-Driven Design: APIs are designed based on user intent (the "Usage Perspective"), decoupled from implementation details.

Emergent Microservices: We do not create empty placeholders. Services emerge physically only after they are defined logically in the schema.

Polyglot by Necessity: Languages are tools selected strictly based on the requirements defined in the contract.

2. The Directory Structure

We utilize a Monorepo-Root Strategy. We avoid nested "App-Level Roots" that create build tool conflicts (especially for Go). The apps/my-app/ directory acts as a logical container, not a build artifact.

/monorepo-root
├── go.work                  # Go Workspace (Maps the entire universe)
├── pom.xml                  # Java Parent (Optional, for shared repo standards)
├── libs
│   └── api-contracts        # THE SOURCE OF TRUTH (Smithy Models)
│       ├── auth.smithy
│       └── cart.smithy
├── apps
│   └── my-app               # LOGICAL CONTAINER
│       ├── gateway          # KrakenD Config (The "Usage" Layer)
│       ├── core             # The "Driver" Service (e.g., Java Monolith)
│       └── image-worker     # Emergent Service (e.g., Go specific task)


3. The "Usage-First" Workflow

This workflow forces architectural discipline before code is written.

Step 1: The Contract (Smithy)

Location: libs/api-contracts/

We design the API based on Usage Perspective. We ask: "What does the user need to accomplish?"

Wrong (Data-Centric): createOrderRow, updateInventoryTable

Right (Usage-Centric): PlaceOrder, CheckItemAvailability

The Separation Trigger: We use Smithy Namespaces and the @service trait to identify distinct capabilities.

// auth.smithy
namespace com.myapp.auth
@service(sdkId: "Auth")
service Authentication { ... }


// payments.smithy
namespace com.myapp.payments
@service(sdkId: "Payments")
service PaymentProcessing { ... }


Outcome: The existence of two @service traits technically mandates two distinct logic flows, validating the need for separation before coding starts.

Step 2: The Aggregation (KrakenD)

Location: apps/my-app/gateway

This layer translates our Backend Truth into the Frontend Experience.

It ingests the OpenAPI/Swagger specs generated from Smithy.

It handles the "Usage Perspective" by merging data.

User Request: GET /dashboard

KrakenD Action: Calls AuthService.getProfile + PaymentService.getRecentTransactions

Response: A single JSON object perfectly shaped for the user.

Step 3: The Implementation (Physical Emergence)

Location: apps/my-app/*

Only now do we choose the language, based on the requirements revealed in Step 1.

Scenario A: AuthService requires complex enterprise logic (OAuth, LDAP).

Decision: Initialize apps/my-app/core using Java.

Scenario B: PaymentService requires high-concurrency webhook handling.

Decision: Initialize apps/my-app/payment-node using Go.

4. Handling the "Empty Space"

Question: How do we start when nothing exists?

Define the Core: Write core.smithy. Define the MVP capabilities.

Initialize the Monolith: Create apps/my-app/core. Implement the Core service interface.

Refactor Later: When core.smithy grows too large, extract a section into video.smithy.

This triggers the creation of apps/my-app/video-worker.

The API Gateway (KrakenD) is updated to route video traffic to the new service.

The User: Notices nothing. The API contract remains stable.

5. Technical Decision Matrix

When analyzing a Smithy Service definition, use this guide to select the implementation language:

Service Characteristic

Recommended Language

Why?

Complex Domain Model



(Banking, Insurance, Inventory)

Java (Spring/Quarkus)

Strongest ecosystem for modeling complex relationships and consistency.

High Throughput / IO



(Streaming, Proxying, Real-time)

Go

Low memory footprint, massive concurrency, "flat" architecture.

Data Intensive / ML



(Recommendations, Analysis)

Python

Unmatched library support for data processing.

BFF (Backend for Frontend)



(If not using KrakenD)

Node.js / TS

JSON-native, shares types with frontend (if TS).

6. Conclusion

This architecture is Contract-Driven and Language-Agnostic.

Smithy defines the What (Capabilities).

KrakenD defines the How (Usage/Delivery).

Polyglot Monorepo defines the Where (Implementation).

By strictly following the Smithy definitions, microservices are not arbitrary choices; they are physical manifestations of the logical contract.