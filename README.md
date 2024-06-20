
# Idempotency Example

This is an example on how to implement idempotency and distributed locking in APIs using [Fiber](https://github.com/gofiber/fiber).


## Run Locally

1. Clone the project

```bash
  git clone https://github.com/wnfrx/idempotency-example.git
```

2. Go to the project directory

```bash
  cd idempotency-example
```

3. Start Redis Server

4. Create .env file

5. Start the server

```bash
  go run main.go
```


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

`REDIS_ADDRESS`

`REDIS_USERNAME`

`REDIS_PASSWORD`

`REDIS_DB`
