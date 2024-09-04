## Go_Backend_Templete

## Overview

This application provides a starting point for comprehensive golang backend. The app is written in Go and utilizes PostgreSQL for the database and Redis for caching.

## Features

- Caching ✅
- authentication ✅
- Async Task ✅
- Emailing ✅
- Database Migration ✅
- Containerization ✅
- Rate Limiting ✅

## Technologies Used

- **Go 1.22**: Backend logic and API
- **PostgreSQL 16**: Relational Database Management
- **Redis 7**: In-memory data structure store for caching

## Libraries Used

- hibbiken/asynq
- Go_redis/cache
- go-redis/redis_rate ->Leaky-bucket rate-limiting with Redis
- pgx driver

## Prerequisites

Ensure you have the following installed on your local machine:

- [Docker](https://www.docker.com/get-started)
- [Go 1.22](https://go.dev/doc/install)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/)

## Getting Started

The follwing startup procedure assumed that you use linux environment and you have Make command installed alredy.
For windows users you may use WSM to archieve the same. Here is the video tutorial on how to set it up.

- [WSM Proper SetUp](https://youtu.be/TtCfDXfSw_0?si=P4FmbpLgY8BCbZ92)

### Clone the Repository

```bash
git clone https://github.com/bstevary/go_backend_templete.git
cd go_backend_templete
```

### Start The backend Automaticaly

If you have the docker compose you need to run the follwing command and backend will be booted and ready to consume trafic on port 8080

```bash
docker compose up
```

alternatively you can consume make command to boot up the app. You must have Docker install

for first time set-up. this will install required services

```bash
make init
```

To start after device boot-up

```bash
make start
```

To resume after os command such close (ctl + c)

```bash
make run
```

### Postman API Docummentation

You can find Postman Collection here

- [API Collection](https://www.postman.com/techgraft/workspace/private-project/collection/23421120-77da6c9e-e182-4da2-be46-5d3ac7f6d14a?action=share&creator=23421120)
