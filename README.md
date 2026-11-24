# LINE Messaging API Emulator

A lightweight, local emulator for the LINE Messaging API that allows developers to test their LINE bot applications without connecting to the actual LINE servers. This emulator implements the endpoints defined in the official LINE OpenAPI specifications, including messaging endpoints and webhook handling.

## Features

- **Local Testing Environment**: Test your LINE bot locally without requiring internet access or LINE accounts
- **Messaging API Coverage**: Implements endpoints from the official LINE Messaging API specification
- **Webhook Support**: Simulates webhook events for testing bot responses
- **No Authentication Required**: Skip the complexity of managing channel access tokens during development
- **OpenAPI Compliant**: Based on the official LINE OpenAPI specifications for accuracy

## Prerequisites

- Go 1.24.4 or higher
- Make (optional, for using Makefile commands)
- Atlas CLI (`curl -sSf https://atlasgo.sh | sh -s --`)

## Installation

Clone the repository:

```bash
git clone https://github.com/zero-color/line-messaging-api-emulator.git
cd line-messaging-api-emulator
```

Install dependencies:

```bash
go mod download
```

## Usage

### Running the Server

Start the emulator server using one of the following methods:

**Using Make:**
```bash
make run
```

**Using Go directly:**
```bash
go run ./cmd/server
```

**With custom port:**
```bash
go run ./cmd/server --port 8080
```

By default, the server runs on port `9090`.

### Available Commands

The project includes a Makefile with helpful commands:

- `make run` - Start the emulator server
- `make test` - Run all tests
- `make lint` - Run code linters
- `make help` - Display all available commands

## API Endpoints

The emulator implements the following LINE Messaging API endpoints:

### Webhook Management
- `GET /v2/bot/channel/webhook/endpoint` - Get webhook endpoint information
- `PUT /v2/bot/channel/webhook/endpoint` - Set webhook endpoint URL
- `POST /v2/bot/channel/webhook/test` - Test webhook endpoint

### Messaging
- `POST /v2/bot/message/reply` - Send reply message
- `POST /v2/bot/message/push` - Send push message
- `POST /v2/bot/message/multicast` - Send multicast message
- `POST /v2/bot/message/narrowcast` - Send narrowcast message
- `POST /v2/bot/message/broadcast` - Send broadcast message

### Content
- `GET /v2/bot/message/{messageId}/content` - Download image, video, and audio data
- `GET /v2/bot/message/{messageId}/content/preview` - Get preview image
- `GET /v2/bot/message/{messageId}/content/transcoding` - Check transcoding status

### User Management
- `GET /v2/bot/profile/{userId}` - Get user profile
- `GET /v2/bot/followers/ids` - Get follower IDs

### Group/Room Management
- `GET /v2/bot/group/{groupId}/member/{userId}` - Get group member profile
- `GET /v2/bot/room/{roomId}/member/{userId}` - Get room member profile
- `POST /v2/bot/group/{groupId}/leave` - Leave group
- `POST /v2/bot/room/{roomId}/leave` - Leave room

### Rich Menu
- `POST /v2/bot/richmenu` / `POST /v2/bot/richmenu/validate` - Create or validate a rich menu
- `GET /v2/bot/richmenu/{richMenuId}` / `GET /v2/bot/richmenu/list` - Get rich menus
- `DELETE /v2/bot/richmenu/{richMenuId}` - Delete a rich menu
- `POST` / `GET /v2/bot/richmenu/{richMenuId}/content` - Upload or download the rich menu image
- `POST` / `DELETE /v2/bot/user/all/richmenu[/{richMenuId}]`, `GET /v2/bot/user/all/richmenu` - Default rich menu
- `POST` / `DELETE` / `GET /v2/bot/user/{userId}/richmenu[/{richMenuId}]` - Per-user rich menu
- `POST /v2/bot/richmenu/bulk/link` / `bulk/unlink` - Link or unlink a rich menu for up to 500 users
- `POST /v2/bot/richmenu/batch`, `validate/batch`, `GET progress/batch` - Batch control (applied synchronously)
- `/v2/bot/richmenu/alias` - Create, update, get, list and delete rich menu aliases

As on the LINE Platform, a rich menu needs an uploaded image before it can be linked or set as default,
and an image can be uploaded only once.

### Quota
- `GET /v2/bot/message/quota` - Always `none` (no limit)
- `GET /v2/bot/message/quota/consumption` - Messages sent this month, counted per recipient; replies are free

### Insight
- `GET /v2/bot/insight/followers` - Current number of followers, for any date
- `GET /v2/bot/insight/demographic` - Always unavailable
- `GET /v2/bot/insight/message/event/aggregation` - Always empty, with null overview values
- `GET /v2/bot/insight/message/delivery`, `GET /v2/bot/insight/message/event` - Always unready or empty

For a complete list of endpoints, refer to the [OpenAPI specification](./line-openapi/messaging-api.yml).

## Admin API

The admin API sets up test data and plays the part of LINE users. It needs no authentication.
See the [OpenAPI specification](./api/adminapi/openapi.yaml) for request and response details.

- `POST /admin/bots` - Create a bot. Its `userId` is both the channel access token and the channel secret.
- `POST /admin/bots/{botId}/followers` - Create dummy users that follow the bot
- `POST /admin/bots/{botId}/users/{userId}/events` - Act as a user: `follow`, `unfollow`, `message` (text) or `postback`.
  The matching webhook event is sent to the bot's webhook endpoint with a valid `X-Line-Signature`,
  and the response reports whether the endpoint accepted it.
- `GET /admin/bots/{botId}/users/{userId}/messages` - List what the bot sent to a user by push, reply,
  multicast, narrowcast or broadcast, with the message objects exactly as the bot sent them

- `GET /admin/bots/{botId}/users/{userId}/rich-menu` - Get the rich menu the user sees: the per-user
  rich menu when one is linked, otherwise the bot's default rich menu

Reply tokens are issued by webhook events, and a reply with a token the emulator didn't issue is rejected.

## Development

### Project Structure

```
.
├── cmd/
│   └── server/         # Server entrypoint
│       └── main.go
├── line-openapi/       # LINE OpenAPI specifications
│   ├── messaging-api.yml
│   ├── webhook.yml
│   └── ...
├── go.mod              # Go module dependencies
├── go.sum              # Go module checksums
├── Makefile            # Build and development commands
└── README.md           # This file
```

### Testing

Run the test suite:

```bash
make test
```

The tests use `gotestsum` for better output formatting and include race condition detection.

### Code Quality

Run linters to ensure code quality:

```bash
make lint
```

The project uses `golangci-lint` for comprehensive code analysis.

## Configuration

The server accepts the following command-line options:

- `--port` - HTTP port number (default: 9090)
- `--database-url` / `DATABASE_URL` - PostgreSQL connection string
- `--skip-schema` / `SKIP_SCHEMA` - Don't create missing tables and indexes at startup

At startup the server creates the tables and indexes that don't exist yet, so an empty database is enough to run it.
It doesn't alter existing tables.
- `--help` - Display help information

Example:
```bash
go run ./cmd/server --port 8080
```

## Integration with Your Bot

To use this emulator with your LINE bot application:

1. Start the emulator server
2. Configure your bot application to use `http://localhost:9090` as the LINE API base URL
3. Point webhook URLs to your bot application
4. Send test messages through the emulator's API endpoints

Example webhook test:
```bash
curl -X POST http://localhost:9090/v2/bot/channel/webhook/test \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "http://localhost:3000/webhook"
  }'
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the terms specified in the [LICENSE](./LICENSE) file.

## Acknowledgments

- Based on the official [LINE OpenAPI specifications](https://github.com/line/line-openapi)
- Built with [chi router](https://github.com/go-chi/chi) for HTTP routing

## Disclaimer

This is an unofficial emulator for development and testing purposes only. It is not affiliated with or endorsed by LINE Corporation. For production use, always use the official LINE Messaging API.

## Support

For issues, questions, or suggestions, please open an issue on the [GitHub repository](https://github.com/zero-color/line-messaging-api-emulator/issues).