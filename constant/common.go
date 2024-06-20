package constant

// Env variables
const (
	EnvRedisAddress  = "REDIS_ADDRESS"
	EnvRedisUsername = "REDIS_USERNAME"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
)

const (
	IdempotencyHeaderKey      = "Idempotency-Key"
	IdempotencyRetryHeaderKey = "Idempotency-Retry"
)
