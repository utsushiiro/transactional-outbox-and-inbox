# Transactional Outbox and Inbox Pattern Example

This repository demonstrates the **Transactional Outbox and Inbox Pattern**, a common technique for ensuring reliable message delivery in distributed systems.

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

