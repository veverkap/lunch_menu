# Copilot Instructions

## Project Overview

This is a Go application that runs daily (via GitHub Actions cron at 2:05 PM UTC) to fetch the next day's school lunch menu and weather, enhance the message with AI, and send it to a Telegram chat. There is also a legacy Ruby version (`runner.rb`) that is no longer actively used.

## Architecture

- **`main.go`** — Application entry point. Orchestrates the daily run: ties together menu fetching, weather fetching, AI enhancement, and Telegram messaging.
- **`menu.go`** — Fetches and parses school lunch menus, and handles reading/writing menu files under `menus/`.
- **`weather.go`** — Fetches and formats weather data from the external weather API.
- **`telegram.go`** — Contains helpers for interacting with the Telegram Bot API (sending messages and photos).
- **`ai.go`** — Builds prompts and calls the AI API to enhance the lunch message content.
- **External APIs used:**
  - [SchoolCafe API](https://webapis.schoolcafe.com) — fetches daily lunch menus by school ID
  - [wttr.in](https://wttr.in) — weather forecast (JSON format)
  - [GitHub Models / OpenAI API](https://models.github.ai/inference) — AI message enhancement via `openai-go` SDK
  - [Telegram Bot API](https://api.telegram.org) — sends messages and photos to a chat
- **`menus/`** — Stores raw menu text files named `YYYY-MM-DD.txt`, committed automatically by CI.
- **`lunch.prompt.yml`** — Reference prompt template for the AI system message (the actual system message is defined as a Go variable `systemMessage` in the Go codebase).

## Build & Run

```bash
go build ./...          # Build
go run main.go          # Run (requires env vars below)
```

There are no tests in this project.

## Required Environment Variables

- `TELEGRAM_TOKEN` — Telegram bot API token
- `TELEGRAM_CHESAPEAKE_CHAT_ID` — Target Telegram chat ID
- `GH_TOKEN` — GitHub token used as the OpenAI API key for GitHub Models

## Key Conventions

- **JSON parsing:** Uses `github.com/buger/jsonparser` for manual JSON traversal rather than unmarshaling into structs. Follow this pattern when adding new API response parsing.
- **Error handling:** Log errors with `logger.Error(...)` and return early. The global `logger` is a `slog.Logger` with JSON output.
- **Telegram methods:** Send JSON payloads via `http.Post` directly (no Telegram SDK). When adding new Telegram Bot API methods, follow the existing `sendTelegramMessage` / `sendTelegramPhoto` pattern.
- **Time zone:** All dates use `America/New_York`. Times are normalized to noon to avoid DST edge cases.
- **CI commits results:** The GitHub Actions workflow runs the program, then auto-commits any new files in `menus/` back to `main`.
