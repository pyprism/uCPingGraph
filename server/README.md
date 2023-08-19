
## Deployment

 - Install [Docker](https://docs.docker.com/engine/installation/)
 - Install [Docker Compose](https://docs.docker.com/compose/install/)
 - Clone this repository
 - Copy .env.example to .env and fill in the values
 - Run `docker compose -f docker-compose.production.yaml up -d` in the server directory

### Commands

 ##### To add new network
`docker compose -f docker-compose.production.yaml exec pgraph /app/main network add`
 ##### To add new device
`docker compose -f docker-compose.production.yaml exec pgraph /app/main device add`
###### Copy the token and paste it in the device's config file 

### Endpoints
###### `POST /api/stats` For pushing data from client to server
Required Header:
```
Authorization: <Token>
```
Body:
```
{
    "latency": 11.2
}
```