# 🚀 Multi-Threaded HTTP Proxy Server

A high-performance, production-ready HTTP proxy server built in Go with concurrent request handling, intelligent caching, and robust error management.

## ✨ Features

- **🔄 Concurrent Request Handling**: Handles multiple client connections simultaneously using goroutines
- **📊 Semaphore-Based Rate Limiting**: Configurable maximum concurrent connections (default: 100)
- **💾 LRU Cache**: In-memory Least Recently Used cache for response caching to improve performance
- **🔒 Thread-Safe Operations**: All cache operations are protected with mutex locks
- **⏱️ Connection Timeouts**: Configurable read/write timeouts for robust network handling
- **🛡️ Graceful Shutdown**: Handles SIGTERM/SIGINT for clean server shutdown
- **📝 Structured Logging**: Separate info and error loggers with timestamps
- **⚙️ Configurable**: Command-line flags and environment variable support

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP Request
       ▼
┌─────────────────────────────────────┐
│     Proxy Server (Port 9000)        │
│  ┌───────────────────────────────┐  │
│  │  Connection Handler           │  │
│  │  - Semaphore (Max 100)        │  │
│  │  - Goroutine per connection   │  │
│  └───────────┬───────────────────┘  │
│              │                       │
│  ┌───────────▼───────────────────┐  │
│  │      LRU Cache               │  │
│  │  - Thread-safe operations    │  │
│  │  - Doubly-linked list        │  │
│  │  - O(1) get/put operations   │  │
│  └───────────┬───────────────────┘  │
│              │                       │
│  ┌───────────▼───────────────────┐  │
│  │   Request Forwarder           │  │
│  │  - TCP-level HTTP forwarding  │  │
│  │  - Response parsing           │  │
│  └───────────┬───────────────────┘  │
└──────────────┼──────────────────────┘
               │ HTTP Request
               ▼
        ┌─────────────┐
        │   Target    │
        │   Server    │
        └─────────────┘
```

## 📂 Project Structure

```
multi-threaded-proxy-server/
├── main.go                 # Entry point and server initialization
├── app.go                  # ProxyServer struct definition
├── proxy.go                # Core proxy logic and request handling
├── config/
│   └── config.go          # Configuration management
├── internal/
│   └── cache/
│       └── lru.go         # LRU cache implementation
├── bin/
│   └── proxy              # Compiled binary
├── go.mod                 # Go module dependencies
└── README.md             # This file
```

## 🚀 Getting Started

### Prerequisites

- Go 1.23.5 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/soumadeep-sarkar/multi-threaded-proxy-server.git
   cd multi-threaded-proxy-server
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the binary**
   ```bash
   go build -o bin/proxy .
   ```

4. **Run the server**
   ```bash
   ./bin/proxy -port=9000
   ```

   Or with custom configuration:
   ```bash
   ./bin/proxy -port=9000 -max-clients=200 -cache-size=500
   ```

### 🐙 Setting Up GitHub Repository

If you want to push this to your own GitHub repository:

1. **Create a new repository on GitHub**
   - Go to https://github.com/new
   - Repository name: `multi-threaded-proxy-server` (or your preferred name)
   - Set it to Public
   - **DO NOT** initialize with README, .gitignore, or license
   - Click "Create repository"

2. **Push your code**
   ```bash
   # Update the remote URL (replace with your username)
   git remote add origin https://github.com/YOUR_USERNAME/multi-threaded-proxy-server.git
   
   # Rename branch to main if needed
   git branch -M main
   
   # Push to GitHub
   git push -u origin main
   ```

   Or use the provided script:
   ```bash
   ./setup-github.sh YOUR_USERNAME multi-threaded-proxy-server
   ```

## ⚙️ Configuration

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `9000` | Port to listen on |
| `-max-clients` | `100` | Maximum concurrent client connections |
| `-cache-size` | `100` | LRU cache capacity |
| `-log-level` | `info` | Log level (info, debug, error) |

### Environment Variables

You can also configure using environment variables:

```bash
export PROXY_PORT=9000
export LOG_LEVEL=info
./bin/proxy
```

## 📖 Usage Examples

### Basic Usage

Start the proxy server:
```bash
./bin/proxy -port=9000
```

### Configure Proxy in Browser

1. **Chrome/Edge**: Settings → Advanced → System → Open proxy settings
2. **Firefox**: Settings → Network Settings → Settings → Manual proxy configuration
3. Set HTTP proxy to `localhost:9000`

### Configure Proxy in curl

```bash
curl -x http://localhost:9000 http://example.com
```

### Configure Proxy in Postman

1. Open Postman Settings
2. Go to Proxy tab
3. Enable proxy and set:
   - Host: `localhost`
   - Port: `9000`

## 🔍 How It Works

1. **Connection Acceptance**: Server listens on specified port and accepts TCP connections
2. **Concurrency Control**: Each connection is handled in a separate goroutine, limited by semaphore
3. **Request Parsing**: HTTP request is parsed from the TCP stream
4. **Cache Lookup**: Server checks LRU cache using key: `host:path:method`
5. **Cache Hit**: If found, cached response is immediately returned
6. **Cache Miss**: Request is forwarded to target server via TCP connection
7. **Response Processing**: Response is parsed, cached, and sent back to client
8. **Resource Cleanup**: Connections are closed and semaphore is released

## 🧪 Testing

### Test with curl

```bash
# Start the proxy server
./bin/proxy -port=9000

# In another terminal, test a request
curl -x http://localhost:9000 http://httpbin.org/get

# Test again (should be cached)
curl -x http://localhost:9000 http://httpbin.org/get
```

### Test with multiple concurrent requests

```bash
# Start the proxy server
./bin/proxy -port=9000

# Test with 10 concurrent requests
for i in {1..10}; do
  curl -x http://localhost:9000 http://httpbin.org/get &
done
wait
```

## 🛠️ Technical Details

### Concurrency Model

- **Goroutine-per-connection**: Each client connection is handled in its own goroutine
- **Semaphore limiting**: Maximum concurrent connections controlled by buffered channel
- **Thread-safe cache**: All cache operations protected with `sync.Mutex`

### Cache Implementation

- **Data Structure**: Doubly-linked list + HashMap for O(1) operations
- **Eviction Policy**: Least Recently Used (LRU)
- **Thread Safety**: All operations are mutex-protected
- **Cache Key Format**: `host:path:method`

### Error Handling

- **Connection Errors**: Logged and gracefully handled
- **Timeout Handling**: Configurable read/write timeouts
- **Panic Recovery**: All goroutines have panic recovery
- **Error Responses**: 502 Bad Gateway for upstream failures

## 📊 Performance Characteristics

- **Concurrent Connections**: Up to 100 (configurable)
- **Cache Operations**: O(1) time complexity
- **Memory Usage**: Proportional to cache size and active connections
- **Throughput**: Limited by network bandwidth and target server response times

## 🔒 Security Considerations

- Currently supports HTTP only (HTTPS support planned)
- No authentication/authorization (suitable for local/trusted networks)
- Input validation on HTTP requests
- Connection timeouts prevent resource exhaustion

## 🚧 Future Enhancements

- [ ] HTTPS/TLS support
- [ ] Cache invalidation policies
- [ ] Request/response filtering
- [ ] Metrics and monitoring endpoints
- [ ] Configuration file support
- [ ] Request logging and analytics
- [ ] Authentication support

## 📝 License

This project is open source and available under the MIT License.

## 👤 Author

**Soumadeep Sarkar**

- GitHub: [@soumadeep-sarkar](https://github.com/soumadeep-sarkar)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the issues page.

## 🙏 Acknowledgments

- Built with Go's excellent concurrency primitives
- Inspired by high-performance proxy server architectures

---

⭐ If you find this project useful, please consider giving it a star!
