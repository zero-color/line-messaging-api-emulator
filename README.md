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
- `POST /v2/bot/richmenu` - Create rich menu
- `GET /v2/bot/richmenu/{richMenuId}` - Get rich menu
- `DELETE /v2/bot/richmenu/{richMenuId}` - Delete rich menu
- `GET /v2/bot/richmenu/list` - Get rich menu list

For a complete list of endpoints, refer to the [OpenAPI specification](./line-openapi/messaging-api.yml).

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