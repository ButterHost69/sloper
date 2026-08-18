Setups Sloper in a Docker Environment.
Add GH Token & Repo Name in the .env file (rename the .env.example -> .env)

## Browser tools for pi workers

The image ships `pi-mcp-adapter` and a chrome-devtools MCP server config
(`/root/.pi/agent/mcp.json`) so pi workers can drive a headless Chrome inside
the container (screenshots, console, network, performance). No extra ports
needed; the browser can reach sloper-web at http://localhost:8080.
