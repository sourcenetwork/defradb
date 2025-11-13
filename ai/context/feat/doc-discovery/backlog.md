# Document Discovery Backlog

## Technical Debt & Improvements

### Performance Optimizations
- **Response Caching**: Cache discovery responses to reduce network overhead for repeated queries
- **Peer Selection Strategy**: Implement intelligent peer selection based on response time and reliability
- **Request Batching**: Batch multiple discovery requests to reduce network chatter
- **Connection Pooling**: Optimize P2P connections for discovery requests

### Error Handling Enhancements
- **Partial Response Handling**: Better handling when only some peers respond
- **Retry Logic**: Exponential backoff for failed discovery requests
- **Circuit Breaker**: Prevent cascading failures when peers are unresponsive
- **Request Tracing**: End-to-end tracing for debugging network issues

### Security Considerations
- **Request Rate Limiting**: Prevent discovery request flooding from malicious peers
- **Access Control Integration**: Respect ACP policies during discovery
- **Query Validation**: Prevent malicious filter expressions
- **Peer Authentication**: Verify peer identity during discovery exchanges

### Code Quality Improvements
- **Interface Abstraction**: Extract discovery protocol interface for testability
- **Configuration Management**: Make timeouts and limits configurable
- **Logging Enhancement**: Structured logging for discovery operations
- **Metrics Collection**: Add telemetry for discovery performance monitoring

## Identified During Development

### Event System Improvements
- **Event Bus Scaling**: Consider dedicated event channels for high-volume discovery requests
- **Response Channel Management**: Proper cleanup of response channels to prevent memory leaks

### P2P Network Optimizations
- **Topic Management**: Evaluate impact of additional discovery topics on network performance
- **Message Serialization**: Consider protobuf for better performance than CBOR

### Testing Gaps
- **Load Testing**: High-volume discovery request testing
- **Network Partition Testing**: Behavior during network splits
- **Performance Benchmarks**: Baseline performance metrics