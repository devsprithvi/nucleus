# Zitadel

> [Zitadel](https://zitadel.com) - Identity infrastructure, simplified.

## Quick Start

```bash
cd service
docker compose up -d
```

Access the console at http://localhost:8080/ui/console

- Login: `zitadel-admin@zitadel.localhost`
- Password: `Password1!`

## Documentation

- 📖 [Zitadel Docs](https://zitadel.com/docs)
- 🐳 [Self-Hosting Guide](https://zitadel.com/docs/self-hosting/deploy/compose)
- 💻 [SDKs](https://zitadel.com/docs/sdk-examples)

## Structure

```
zitadel/
└── service/
    ├── docker-compose.yaml   # Zitadel + PostgreSQL
    └── README.md
```
