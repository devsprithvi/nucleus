# {Module Display Name}

> {Brief one-line description of the module.}

## Overview

{2-3 sentences explaining what this module does and why it exists.}

## Installation

```bash
go get nucleus/modules/{domain}/{module-name}
```

## Quick Start

{Flexible section — adapt based on what makes sense for this module.}
{For service-backed modules: show how to call the service.}
{For library modules: show import and basic usage.}

```go
import "nucleus/modules/{domain}/{module-name}"

// Example usage
{code example here}
```

## What This Module Provides

- ✅ {Feature 1}
- ✅ {Feature 2}
- ✅ {Feature 3}

## Powered By

{Remove this section if the module is original code.}

This module is powered by **{Tool Name}**. For advanced configuration:

- 📖 [{Tool Name} Documentation]({docs-url})
- 🔧 [{Tool Name} SDK]({sdk-url})

## Service

{Remove this section if no microservice is provided.}

This module also exposes a microservice:

| Protocol | Endpoint            |
| -------- | ------------------- |
| gRPC     | `{service.v1}`      |
| REST     | `/api/v1/{service}` |

### Using the Service Directly

{Show how to call the service if that's the recommended approach.}

## Configuration

{Remove if no configuration is needed.}

| Environment Variable | Description   | Default     |
| -------------------- | ------------- | ----------- |
| `{VAR_NAME}`         | {Description} | `{default}` |

## Examples

{Point to examples folder if it exists, or remove this section.}

See the [examples](./examples/) folder for more usage patterns.

## Dependencies

This module depends on:

- [`observability/logging`](../../observability/logging/) — Structured logging
- [`observability/metrics`](../../observability/metrics/) — Metrics collection

---

_See [module.yaml](./module.yaml) for machine-readable metadata._
