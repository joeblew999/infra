# Issue: goctl should generate message queue code for KQ and DQ

## Summary
goctl provides comprehensive code generation for APIs, RPCs, models, Docker, and Kubernetes, but lacks code generation for message queues despite go-zero having a mature messaging framework (go-queue). This forces developers to manually write repetitive boilerplate code for producers, consumers, and queue configurations.

## Current Behavior
Developers must manually create all message queue code:

❌ **Manual setup required for:**
- Queue configuration structs
- Producer service code
- Consumer handler implementations  
- Message type definitions
- Service registration and startup
- Directory structure organization

**Example manual work from go-queue examples:**
```go
// Manual configuration struct
type Config struct {
    rest.RestConf
    KqPusherConf kq.KqPusherConf
    KqConsumerConf kq.KqConf
}

// Manual consumer setup
q := kq.MustNewQueue(c.KqConsumerConf, kq.WithHandle(func(ctx context.Context, k, v string) error {
    // Manual business logic for each message
    fmt.Printf("Processing payment: %s\n", v)
    return nil
}))
defer q.Stop()
q.Start()

// Manual producer setup  
pusher := kq.NewPusher([]string{"127.0.0.1:9092"}, "payment-topic")
```

## Expected Behavior
goctl should generate message queue code similar to how it generates API and RPC services.

## Proposed Solution
Add queue code generation commands to goctl:

```bash
# Generate Kafka queue services
goctl queue kafka -config queue.yaml -dir .

# Generate NATS queue services
goctl queue nats -config queue.yaml -dir .

# Generate delay queue services  
goctl queue delay -config delay.yaml -dir .

# Generate with specific patterns
goctl queue kafka -config queue.yaml -dir . --with-producers --with-consumers
goctl queue nats -config queue.yaml -dir . --with-pub --with-sub
```

## Suggested Generated Structure

### Configuration File Format
`queue.yaml`:
```yaml
name: payment-service
queues:
  - name: payment-success
    type: kafka
    producers: [order-service]
    consumers: [notification-service, audit-service]
  - name: user-events
    type: nats
    subjects: [user.created, user.updated]
    publishers: [user-service]
    subscribers: [email-service, analytics-service]
  - name: payment-retry
    type: delay
    delay: 5m
    consumers: [retry-service]
```

### Generated Files

goctl should generate the standard queue service structure:

- **Configuration files** - Queue connection and topic configuration
- **Producer handlers** - Service methods to publish messages  
- **Consumer handlers** - Message processing logic with TODO placeholders
- **Service registration** - Auto-wire queues into the service lifecycle
- **Directory structure** - Consistent with API and RPC service patterns

## Why This Matters

1. **Consistency** - All other go-zero components have code generation
2. **Developer Productivity** - Eliminates repetitive queue boilerplate
3. **Best Practices** - Enforces consistent queue service patterns
4. **Complete Solution** - go-zero should be fully code-generation driven
5. **Framework Parity** - Other frameworks (gRPC, etc.) generate message handling code

## Current Workaround Gap

The go-queue examples show clear, repetitive patterns that are perfect for code generation:
- Configuration struct patterns
- Producer/consumer interface implementations
- Service registration boilerplate
- Directory structure conventions

## Benefits of Queue Code Generation

- **Faster development** with instant queue service scaffolding
- **Reduced errors** from manual boilerplate mistakes
- **Consistent architecture** across queue-based services
- **Better onboarding** for developers new to go-zero messaging
- **Complete go-zero experience** - generate APIs, queues, and deployment together

## Integration with Existing Commands

Queue generation should follow existing goctl command patterns:

```bash
# New top-level queue command (similar to api, rpc, model)
goctl queue kafka -config queues.yaml -dir .
goctl queue nats -config queues.yaml -dir .
goctl queue delay -config queues.yaml -dir .

# Follow the same pattern as:
# goctl api go -api user.api -dir .
# goctl rpc new user-rpc
# goctl model mysql datasource -table users -dir .
```

## References
- [go-queue Examples](https://github.com/zeromicro/go-queue)
- [go-zero Message Queue Docs](https://go-zero.dev/en/docs/tutorials/message-queue/kafka)
- [go-zero Delay Queue Docs](https://go-zero.dev/en/docs/tutorials/delay-queue/beanstalkd)

## Version Info
- goctl version: 1.8.5
- go-zero version: 1.9.0
- go-queue version: 1.5.3

Adding message queue code generation would complete go-zero's vision of comprehensive microservice development through code generation, making it truly competitive with other full-stack microservice frameworks.

## Repository
Submit this issue to: https://github.com/zeromicro/go-zero/issues/new