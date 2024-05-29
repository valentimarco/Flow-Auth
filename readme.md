# Flow-Auth
Microservice to handle all the auth system:

- provide a simple RESTAPI to authenticate and authorize users
- Fack Javascript. Long life to golang!
    ![alt text](image.png)
Based on https://dev.to/egregors/passkey-in-go-1efk
## Dev guide

1. install air for hotreload watcher:
   ```bash
    go install github.com/cosmtrek/air@latest
   ```
2. install dependencies
   ```bash
   go mod download
   ```
3. run the watcher:
   ```bash
   air
   ```
## Docker guide

1. `docker compose build`
2. `docker compose up`
3. Enjoy