# Transactional Outbox and Inbox Pattern Example

This repository demonstrates the **Transactional Outbox and Inbox Pattern**, a common technique for ensuring reliable message delivery in distributed systems. For a detailed explanation of this pattern, the article [Microservices 101: Transactional Outbox and Inbox](https://softwaremill.com/microservices-101/) on the SoftwareMill blog would be helpful.

## How to Use

Follow the steps below to build and run the example.

### 1. Build the Application

Run the following command to build the application:

```bash
$ make build
```

---

### 2. Run the Application

Start the application using Docker Compose to set up the required environment:

```bash
$ docker compose up
```

Next, start the producer and consumer services. The producer generates messages, and the consumer processes them:

1. Start the producer:

```bash
$ ./bin/server producer
```

2. Start the consumer:

```bash
$ ./bin/server consumer
```

## Notes

This example is designed to run locally and is not intended for production use without further enhancements. 

Some possible enhancements for real-world systems include:
- Monitoring and Alerting: Ensure the background worker is healthy and functioning correctly by collecting metrics (e.g., number of events processed, failure rates) and setting up alerts for issues like repeated failures or large backlogs in the outbox/inbox table.
- Optimizations:
	- Batching: Process events in batches (e.g., 100 at a time) to reduce overhead and improve efficiency.\*
	- Parallelism: Run multiple worker instances in parallel to handle a high throughput of events and improve scalability.
- Failover and High Availability: Deploy multiple worker instances to ensure high availability, allowing one instance to take over if another crashes.

\* Batching is planned to be implemented in a future version of this example.
