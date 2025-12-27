package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/soumadeep-sarkar/multi-threaded-proxy-server/internal/cache"
)

const (
	badGatewayBody = "The target server is unavailable. Please try again later."
	readTimeout    = 30 * time.Second
	writeTimeout   = 30 * time.Second
)

var badGatewayResponse = fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\n"+
	"Content-Type: text/plain\r\n"+
	"Content-Length: %d\r\n"+
	"Connection: close\r\n"+
	"\r\n"+
	badGatewayBody, len(badGatewayBody))

// handleConnection processes a client connection
func (s *ProxyServer) handleConnection(conn net.Conn) {
	// Acquire semaphore to limit concurrent connections
	s.semaphore <- struct{}{}
	defer func() {
		<-s.semaphore
	}()

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			s.errorLogger.Printf("Recovered from panic: %v", r)
		}
	}()

	// Set timeouts
	conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	defer conn.Close()

	// Parse HTTP request
	reader := bufio.NewReader(conn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		s.errorLogger.Printf("Failed to parse HTTP request: %v", err)
		return
	}

	// Log request
	s.infoLogger.Printf("Received %s request for %s%s", request.Method, request.Host, request.URL.Path)

	// Check cache
	cacheKey := s.generateCacheKey(request)
	if cachedItem := s.getCachedResponse(cacheKey); cachedItem != nil {
		s.sendCachedResponse(conn, cachedItem)
		return
	}

	// Forward request to target server
	response, err := s.forwardRequest(request)
	if err != nil {
		s.errorLogger.Printf("Failed to forward request: %v", err)
		s.sendErrorResponse(conn)
		return
	}
	defer response.Body.Close()

	// Prepare and send response
	statusLine, headers, body := s.prepareResponse(response)
	responseData := statusLine + headers + "\r\n" + string(body)

	if _, err := conn.Write([]byte(responseData)); err != nil {
		s.errorLogger.Printf("Failed to send response: %v", err)
		return
	}

	// Cache the response
	s.cacheResponse(cacheKey, statusLine, headers, body)
	s.infoLogger.Printf("Successfully forwarded and cached response for %s%s", request.Host, request.URL.Path)
}

// generateCacheKey creates a unique cache key from the request
func (s *ProxyServer) generateCacheKey(req *http.Request) string {
	return fmt.Sprintf("%s:%s:%s", req.Host, req.URL.Path, req.Method)
}

// getCachedResponse retrieves a response from cache if available
func (s *ProxyServer) getCachedResponse(key string) *cache.ProxyItem {
	found, item := s.cache.Get(key)
	if found && item != nil {
		s.infoLogger.Printf("Cache HIT for key: %s", key)
		return item
	}
	s.infoLogger.Printf("Cache MISS for key: %s", key)
	return nil
}

// sendCachedResponse sends a cached response to the client
func (s *ProxyServer) sendCachedResponse(conn net.Conn, item *cache.ProxyItem) {
	response := item.Status + item.Header + "\r\n" + string(item.Body)
	if _, err := conn.Write([]byte(response)); err != nil {
		s.errorLogger.Printf("Failed to send cached response: %v", err)
	}
}

// sendErrorResponse sends a 502 Bad Gateway error response
func (s *ProxyServer) sendErrorResponse(conn net.Conn) {
	if _, err := conn.Write([]byte(badGatewayResponse)); err != nil {
		s.errorLogger.Printf("Failed to send error response: %v", err)
	}
}

// forwardRequest forwards the HTTP request to the target server
func (s *ProxyServer) forwardRequest(req *http.Request) (*http.Response, error) {
	// Determine target address
	targetAddr := req.Host
	if !strings.Contains(targetAddr, ":") {
		targetAddr += ":80"
	}

	// Establish TCP connection to target server
	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	// Set timeouts
	targetConn.SetDeadline(time.Now().Add(30 * time.Second))

	// Build HTTP request string
	requestLine := fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.Path, req.Proto)
	hostHeader := fmt.Sprintf("Host: %s\r\n", req.Host)

	// Build headers
	var headers strings.Builder
	for key, values := range req.Header {
		headers.WriteString(fmt.Sprintf("%s: %s\r\n", key, strings.Join(values, ", ")))
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Construct full request
	fullRequest := requestLine + hostHeader + headers.String() + "\r\n" + string(body)

	// Send request to target server
	if _, err := targetConn.Write([]byte(fullRequest)); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response from target server
	reader := bufio.NewReader(targetConn)
	response, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}

// prepareResponse extracts status, headers, and body from HTTP response
func (s *ProxyServer) prepareResponse(resp *http.Response) (string, string, []byte) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.errorLogger.Printf("Failed to read response body: %v", err)
		body = []byte{}
	}

	// Build headers string
	var headers strings.Builder
	for key, values := range resp.Header {
		headers.WriteString(fmt.Sprintf("%s: %s\r\n", key, strings.Join(values, ", ")))
	}

	// Build status line
	statusLine := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

	return statusLine, headers.String(), body
}

// cacheResponse stores the response in the cache
func (s *ProxyServer) cacheResponse(key, status, headers string, body []byte) {
	item := &cache.ProxyItem{
		Status: status,
		Header: headers,
		Body:   body,
	}
	s.cache.Put(key, item)
}
