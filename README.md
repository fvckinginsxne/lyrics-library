# Lyrics Library API

RESTful microservice in Go for receiving song lyrics with Russian translation.

## Features
- Save new lyrics with translation by artist and title
- Get song lyrics by artist and track title
- Delete lyrics by UUID
- Automatic translation into Russian

## Stack
- **Language**: Go 1.24+
- **Web framework**: Gin
- **Logging**: log/slog
- **Database**: PostgreSQL
- **Migrations**: golang-migrate
- **Caching**: Redis
- **Containerization**: Docker
- **External APIs**:
  - [LyricsOVH](https://lyricsovh.docs.apiary.io/#reference) - fetching lyrics
  - [Yandex.Translate](https://yandex.cloud/ru/docs/translate/quickstart) - translation into Russian
- **Documentation**: Swagger

## Quick Start
### 1. Clone Repository
```
git clone https://github.com/fvckinginsxne/lyrics-library.git
cd app
```
### 2. Setup environment
```
cp .env.example .env
nano .env 
```
### 3. Start application
```
docker-compose --env-file .env up -d
```
### 4. The documentation is located at
```
localhost:8080/swagger/index.html
```

## TODO 
- [ ] Tests
- [ ] Add integration with auth service using gRPC  
- [ ] Use kafka/rabbitmq
- [ ] Make frontend
- [ ] Deploy fullstack app on server
