# Zitadel on Hugging Face Spaces

> ⚠️ **Temporary Solution** - All-in-one container for development/testing only.

## What This Does

Deploys Zitadel to Hugging Face Spaces using a single Docker container that includes:

- PostgreSQL database
- Zitadel auth server

## How to Deploy

### Step 1: Push to Hugging Face

```bash
# Create a new HF Space
huggingface-cli repo create nucleus-zitadel --type space --space_sdk docker

# Clone it
git clone https://huggingface.co/spaces/YOUR_USERNAME/nucleus-zitadel
cd nucleus-zitadel

# Copy the Dockerfile
cp /path/to/Dockerfile .

# Push
git add .
git commit -m "Initial Zitadel deployment"
git push
```

### Step 2: Or Use Terraform

```bash
cd terraform
terraform init
terraform apply
```

Then manually push the Dockerfile to the created Space.

## Default Credentials

| Field      | Value                                    |
| ---------- | ---------------------------------------- |
| URL        | `https://nucleus-zitadel.hf.space`       |
| Admin User | `admin@zitadel.nucleus-zitadel.hf.space` |
| Password   | `Admin123!`                              |

## Limitations

- **Not production-ready** - Database in same container
- **No persistent storage** - Data lost on container restart (unless HF persistent storage enabled)
- **Resource intensive** - Needs `cpu-upgrade` tier minimum

## For Production

Use the original `docker-compose.yaml` with:

- Proper separated services
- External managed database
- Or use [Zitadel Cloud](https://zitadel.cloud)
