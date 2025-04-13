# Lyrics Library API

RESTful microservice in Go for receiving song lyrics with Russian translation.

## Features
- Getting song lyrics by artist and track title
- Automatic translation into Russian

## Stack
- **Language**: Go 1.24+
- **Database**: PostgreSQL
- **Caching**: Redis
- **Migrations**: golang-migrate
- **External APIs**:
  - [LyricsOVH](https://lyricsovh.docs.apiary.io/#reference) - fetching lyrics
  - [Yandex.Translate](https://yandex.cloud/ru/docs/translate/quickstart) - translation into Russian
- **Containerization**: Docker

## Quick Start
### 1. Clone Repository
```bash
git clone https://github.com/fvckinginsxne/lyrics-library.git
cd lyrics-library
```
### 2. Setup environment
```bash
cp .env.example .env
nano .env 
```
### 3. Start services (Postgres, Redis)
```bash
docker-compose --env-file .env up -d
```
### 4. Apply database migrations
```bash
CONFIG_PATH=.env go run ./cmd/migrator/main.go --migrations-path=./migrations --action=up --force-version=0
```
### 5. Run application
```bash
go run ./cmd/lyrics-library --config=.env
```

## TODO 
- [ ] Tests
- [ ] Add integration with auth service using gRPC  
- [ ] Use kafka/rabbitmq
- [ ] Make frontend
- [ ] Deploy fullstack app on server
