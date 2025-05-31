# message-sender

Every 2 minutes, read database and send the unsent messages next in line.

### Build and run with Docker

To start the app:

```sh
docker compose up
```

- The app will be available at http://localhost:8080
    - http://localhost:8080/sent-messages - Get sent messages
    - http://localhost:8080/change-state?action=pause - Pause the message sender
    - http://localhost:8080/change-state?action=resume - Resume the message sender
    (You can also use any [OpenAPI UI](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/taylankasap/message-sender/refs/heads/master/api/openapi.yaml) to see the endpoints)
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

### Possible improvements

These are possible improvements that could've been done if this was a production app:

- Use another database on a remote server
- Create multiple config files for different environments instead of giving everything in main.go
- Create golangci-lint config to make it stay consistent among updates
- Check the 3rd party API response for errors instead of always setting the message status to ‘sent’
- Implement retries with exponential backoff in MessageDispatcher
- Pass logger around instead of using global logger
- Add pagination to get sent messages endpoint
- Watch db for changes and send a message if within 2 minute rate limit, rather than checking every 2 minutes (which causes some delay for new messages)
- Add message character limit to the database too
- Handle edge cases such as message is sent but could not be marked as sent
- CI/CD pipeline to run tests and linter on every commit
