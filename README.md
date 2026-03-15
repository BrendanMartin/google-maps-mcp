# Google Maps MCP Server

A lightweight [MCP](https://modelcontextprotocol.io/) server that wraps Google Maps REST APIs over stdio. Built in Go for use with Claude Code and other MCP-compatible clients.

## Tools

| Tool | Description |
|------|-------------|
| `maps_geocode` | Convert an address to latitude/longitude coordinates |
| `maps_reverse_geocode` | Convert coordinates to an address |
| `maps_directions` | Get directions with step-by-step navigation |
| `maps_distance_matrix` | Calculate distances and travel times between multiple origins/destinations |
| `maps_search_places` | Search for places by text query |

## Setup

### Prerequisites

- Go 1.24+
- A [Google Maps API key](https://developers.google.com/maps/documentation/geocoding/get-api-key)

### Build

```bash
go build -o google-maps-mcp.exe .
```

### Configure in Claude Code

Add to `~/.claude.json` under `mcpServers`:

```json
{
  "mcpServers": {
    "google-maps": {
      "type": "stdio",
      "command": "/path/to/google-maps-mcp",
      "env": {
        "GOOGLE_MAPS_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Usage

Once configured, Claude Code will have access to the maps tools. Example prompts:

- "How far is it from NYC to Boston?"
- "Find coffee shops near Times Square"
- "What's the address at 40.7128, -74.0060?"
- "Get directions from LAX to Santa Monica by transit"

## License

MIT
