# Sloper Docker Setup

Setups Sloper in a Docker Environment.
Add GH Token & Repo Name in the .env file (rename the .env.example -> .env)

Each container also runs the read-only **dashboard API** (`sloper-web`) on port
`8080` so the [console](../dashboard/README.md) can observe the instance:

```bash
make docker              # builds sloper + sloper-web image and starts it
# dashboard API is now reachable at http://localhost:8080/api/health
```

To change the host port, set `SLOPER_WEB_PORT` in `.env` (e.g. `8081` for a
second instance) — it is forwarded by `compose.yml`.