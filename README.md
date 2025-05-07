## URL Shortener

This project provides a test service for shortening URLs with a database and caching system. 

The system has a basic rate limiter that doesn't allow requests to be made more than **once per second from one IP address**. The cache has a lifetime and is automatically cleared to save memory.

The system uses **base62** encoding to turn the primary key into a unique string, which reduces the amount of data to store and makes searching easier.

System includes three main handlers:

1. **Shorten**: for creating a shortened URL.
2. **Redirect**: for redirecting to the original URL based on the shortened URL.
3. **Stats**: for retrieving statistics on the shortened URL.



### Running

The project uses Docker to run and has a Docker Compose file and a Dockerfile that define all the necessary steps for building and running the project. Multi-stage Building is used to reduce the image size.

```bash
docker-compose up -d
```

### API Handlers

#### 1. Shorten (POST /shorten)

This handler accepts the original URL and creates a shortened version of it.

**Request:**

```bash
curl -X POST http://localhost:8080/shorten \
-H "Content-Type: application/json" \
-d '{"original": "https://example.com"}'
```

Response:

```json
{
  "original": "https://example.com",
  "short": "AbC",
  "clicks": 0
}
```
#### 2. Redirect (GET /:short)
   This handler redirects to the original URL based on the shortened identifier.

Request:
```bash 
curl -i http://localhost:8080/AbC 
```
Response:
```https
https/1.1 302 Found
Location: https://example.com
```
If the shortened URL doesn't exist, the server will return **404 Not Found**.

#### 3. Stats (GET /stats/:short)
   This handler provides statistics for the shortened URL, including the number of clicks.

Request:

```bash 
curl http://localhost:8080/stats/AbC
```
Response:
```json
{
  "original": "https://example.com",
  "short": "AbC",
  "clicks": 10
}
```
If the shortened URL doesn't exist, the server will return **404 Not Found**.

