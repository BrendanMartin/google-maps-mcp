# Google Maps MCP Server

A lightweight [MCP](https://modelcontextprotocol.io/) server that wraps Google Maps REST APIs over stdio. Built in Go for use with Claude Code and other MCP-compatible clients. No SDK or abstraction layer — just direct HTTP calls to the Google Maps API with results formatted as plain text.

## Why this over the Node.js/Docker alternatives?

The [official MCP Google Maps server](https://github.com/modelcontextprotocol/servers-archived/tree/main/src/google-maps) (now archived) and community alternatives all require Node.js or Docker. This one is a single static binary (~11MB) with zero runtime dependencies — no `node_modules`, no container, no `npx`. It starts instantly and uses minimal memory.

| | This server | Node.js alternatives |
|---|---|---|
| Runtime | None (static binary) | Node.js + npm |
| Install | Download binary | `npx` or Docker |
| Binary size | ~11MB | ~100MB+ (with node_modules) |
| Startup | Instant | ~1-2s (Node.js cold start) |
| Dependencies | 0 | npm packages |

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

- A [Google Maps API key](https://developers.google.com/maps/documentation/geocoding/get-api-key)

### Install

Download the pre-built binary for your platform from [Releases](https://github.com/BrendanMartin/google-maps-mcp/releases), or build from source:

```bash
# Option 1: go install
go install github.com/brendanmartin/google-maps-mcp@latest

# Option 2: build from source
git clone https://github.com/BrendanMartin/google-maps-mcp.git
cd google-maps-mcp
go build -o google-maps-mcp .
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

## Example: Claude Code Agent

You can create a dedicated maps agent that runs on a lightweight model. Save this as `~/.claude/agents/maps.md`:

```markdown
---
name: maps
description: Location and maps agent for geocoding, directions, distances, and place search
model: haiku
tools: mcp__google-maps__maps_geocode, mcp__google-maps__maps_reverse_geocode, mcp__google-maps__maps_directions, mcp__google-maps__maps_distance_matrix, mcp__google-maps__maps_search_places, WebSearch, WebFetch
---

You are a location and maps assistant. Use the Google Maps MCP tools to answer questions about places, directions, distances, and coordinates.

## Available tools

- **maps_geocode** — Convert an address to lat/lng coordinates
- **maps_reverse_geocode** — Convert lat/lng to an address
- **maps_directions** — Get directions between two locations (supports driving, walking, bicycling, transit)
- **maps_distance_matrix** — Calculate distances and travel times between multiple origins and destinations
- **maps_search_places** — Search for places by text query (e.g. "coffee shops near Central Park")

## Guidelines

1. Use the most specific tool for the task
2. For directions, default to driving mode unless the user specifies otherwise
3. For distance_matrix, batch multiple origins/destinations into a single call when possible
4. Use WebSearch/WebFetch to supplement maps data with additional context when helpful (hours, reviews, menus, etc.)
5. Return concise, well-formatted results
6. When comparing options, present results in a clear table or list

Multiple instances of this agent can be launched in parallel for independent location queries.
```

Then add a pointer in your `CLAUDE.md`:

```markdown
## Maps & Location
When the user asks about directions, distances, travel times, place search, geocoding, or anything
location-related, use the `maps` agent which runs on Haiku. Spawn multiple agents in parallel for
independent location queries.
```

## License

MIT
