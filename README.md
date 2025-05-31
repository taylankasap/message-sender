# message-sender

Every 2 minutes, read database and send the next unsent message.

### Build and run with Docker

To start the app:

```sh
docker compose up
```

- The app will be available at http://localhost:8080
- The SQLite database will be persisted in `data/db.sqlite3`. The app will seed the database on first start-up.
- You can list the keys in Redis with: `docker compose exec -it redis redis-cli KEYS '*'`

### How to start the app

If you want to start the app this way, you may want to setup Redis (not required, you just will see logs in the console).

```
make run
```

### How to test the app

```
make test
```

### How to run the linter

```
make lint
```

This command will sort imports consistently, fix whitespaces around the code and show errors in case of any other lint issues.

### Notes

We're using OpenAPI 3.0.0 instead of 3.1.0 because oapi-codegen currently does not support 3.1.0.
