# Lyrics Library API

RESTful microservice in Go for receiving song lyrics with Russian translation.

## Features
- Getting song lyrics by artist and track title
- Automatic translation into Russian

## Stack
- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Caching**: Redis
- **Containerization**: Docker
- **External APIs**:
  - [LyricsOVH](https://lyricsovh.docs.apiary.io/#reference) - fetching lyrics
  - [Yandex.Translate](https://yandex.cloud/ru/docs/translate/quickstart) - translation into Russian

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
### 3. Start application
```bash
docker-compose --env-file .env up -d
```

## TODO 
- [ ] Tests
- [ ] Add integration with auth service using gRPC  
- [ ] Use kafka/rabbitmq
- [ ] Make frontend
- [ ] Deploy fullstack app on server
