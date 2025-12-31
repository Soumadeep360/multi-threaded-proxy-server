# 🏗️ Architecture Documentation

## System Architecture

### High-Level Overview

The Multi-Threaded Proxy Server is designed as a forward HTTP proxy that sits between clients and target servers, providing caching and concurrent request handling capabilities.

```mermaid
graph TB
    Client[Client Application] -->|HTTP Request| Proxy[Proxy Server<br/>Port 9000]
    Proxy -->|Check Cache| Cache[LRU Cache]
    Cache -->|Cache Hit| Proxy
    Proxy -->|Cache Miss| Forwarder[Request Forwarder]
    Forwarder -->|HTTP Request| Target[Target Server]
    Target -->|HTTP Response| Forwarder
    Forwarder -->|Store & Return| Proxy
    Proxy -->|HTTP Response| Client
    
    style Proxy fill:#4CAF50
    style Cache fill:#2196F3
    style Forwarder fill:#FF9800
```

## Component Architecture

### 1. Main Server Component

```mermaid
classDiagram
    class ProxyServer {
        -infoLogger *log.Logger
        -errorLogger *log.Logger
        -cache *LRUCache
        -config *Config
        -semaphore chan struct{}
        +Start() error
        +handleConnection(conn net.Conn)
        +forwardRequest(req *http.Request) *http.Response
        +prepareResponse(resp *http.Response) string, string, []byte
    }
    
    class Config {
        +Port string
        +MaxClients int
        +CacheSize int
    }
    
    class LRUCache {
        -mu sync.Mutex
        -capacity int
        -size int
        -head *node
        -tail *node
        -items map[string]*node
        +Get(key string) bool, *ProxyItem
        +Put(key string, value *ProxyItem)
        +Size() int
        +Capacity() int
    }
    
    ProxyServer --> Config
    ProxyServer --> LRUCache
```

### 2. Request Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy Server
    participant S as Semaphore
    participant Cache as LRU Cache
    participant T as Target Server
    
    C->>P: TCP Connection
    P->>S: Acquire Semaphore
    S-->>P: Semaphore Acquired
    
    C->>P: HTTP Request
    P->>Cache: Get(cacheKey)
    
    alt Cache Hit
        Cache-->>P: Cached Response
        P->>C: HTTP Response (from cache)
    else Cache Miss
        Cache-->>P: Not Found
        P->>T: Forward HTTP Request
        T-->>P: HTTP Response
        P->>Cache: Put(cacheKey, response)
        P->>C: HTTP Response
    end
    
    P->>S: Release Semaphore
    P->>C: Close Connection
```

### 3. Concurrency Model

```mermaid
graph LR
    Listener[TCP Listener] -->|Accept| Conn1[Connection 1]
    Listener -->|Accept| Conn2[Connection 2]
    Listener -->|Accept| ConnN[Connection N]
    
    Conn1 -->|Goroutine| Handler1[Handler 1]
    Conn2 -->|Goroutine| Handler2[Handler 2]
    ConnN -->|Goroutine| HandlerN[Handler N]
    
    Handler1 --> Semaphore[Semaphore<br/>Max: 100]
    Handler2 --> Semaphore
    HandlerN --> Semaphore
    
    Semaphore --> Cache[Shared LRU Cache<br/>Thread-Safe]
```

### 4. Cache Data Structure

```mermaid
graph LR
    subgraph "LRU Cache Structure"
        Map[HashMap<br/>O(1) Lookup] --> Node1[Node 1<br/>Key: host:path:method<br/>Value: ProxyItem]
        Map --> Node2[Node 2]
        Map --> NodeN[Node N]
        
        Node1 <-->|prev| Node2
        Node2 <-->|next| Node1
        Node2 <-->|prev| NodeN
        NodeN <-->|next| Node2
        
        Head[Head<br/>Most Recent] --> Node1
        Tail[Tail<br/>Least Recent] --> NodeN
    end
    
    style Head fill:#4CAF50
    style Tail fill:#F44336
```

## Data Flow

### Request Processing Pipeline

```
1. TCP Connection Accepted
   ↓
2. Semaphore Acquired (if available)
   ↓
3. HTTP Request Parsed
   ↓
4. Cache Key Generated: "host:path:method"
   ↓
5. Cache Lookup
   ├─→ Cache Hit: Return cached response
   └─→ Cache Miss: Continue to step 6
   ↓
6. Forward Request to Target Server
   ├─→ Establish TCP connection
   ├─→ Build HTTP request string
   ├─→ Send request
   └─→ Read response
   ↓
7. Parse Response
   ├─→ Extract status line
   ├─→ Extract headers
   └─→ Extract body
   ↓
8. Store in Cache (if cacheable)
   ↓
9. Send Response to Client
   ↓
10. Release Semaphore & Close Connections
```

## Thread Safety

### Mutex Protection

All cache operations are protected by a `sync.Mutex`:

```go
func (c *LRUCache) Get(key string) (bool, *ProxyItem) {
    c.mu.Lock()         // Acquire lock
    defer c.mu.Unlock() // Release lock on return
    
    // ... cache operations
}
```

### Semaphore-Based Concurrency Control

```go
// Semaphore limits concurrent connections
semaphore := make(chan struct{}, maxClients)

// In handler:
semaphore <- struct{}{}  // Acquire
defer func() {
    <-semaphore  // Release
}()
```

## Error Handling Strategy

```mermaid
graph TD
    Error[Error Occurs] --> Type{Error Type}
    
    Type -->|Connection Error| Log[Log Error]
    Type -->|Parse Error| Log
    Type -->|Timeout Error| Log
    
    Log --> Response{Can Send<br/>Response?}
    
    Response -->|Yes| Send502[Send 502 Bad Gateway]
    Response -->|No| Close[Close Connection]
    
    Send502 --> Close
    Close --> Cleanup[Cleanup Resources]
    Cleanup --> End[End Handler]
```

## Performance Optimizations

1. **Connection Pooling**: Reuse TCP connections where possible
2. **LRU Caching**: Fast O(1) cache operations
3. **Goroutine Pooling**: Efficient concurrent request handling
4. **Zero-Copy Where Possible**: Minimize memory allocations
5. **Timeout Management**: Prevent resource leaks

## Scalability Considerations

- **Horizontal Scaling**: Can run multiple instances behind a load balancer
- **Vertical Scaling**: Limited by Go's goroutine efficiency (can handle thousands)
- **Cache Scaling**: In-memory cache limits to single instance; distributed cache would require external solution
- **Connection Limits**: Configurable via `-max-clients` flag

## Security Architecture

```
┌─────────────────────────────────────┐
│         Security Layers             │
├─────────────────────────────────────┤
│ 1. Connection Timeouts              │
│ 2. Input Validation                 │
│ 3. Resource Limits (Semaphore)      │
│ 4. Panic Recovery                   │
│ 5. Error Sanitization               │
└─────────────────────────────────────┘
```

## Monitoring Points

1. **Connection Count**: Track active connections
2. **Cache Hit Rate**: Monitor cache effectiveness
3. **Request Latency**: Measure response times
4. **Error Rate**: Track failed requests
5. **Memory Usage**: Monitor cache size and growth

---

*This architecture is designed for high performance, reliability, and maintainability.*

