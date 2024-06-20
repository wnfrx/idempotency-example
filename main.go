package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"github.com/wnfrx/idempotency-example/constant"
	"github.com/wnfrx/idempotency-example/model"
	"github.com/wnfrx/idempotency-example/util"
)

var counter = model.Counter{}
var redisClient = redis.NewClient(&redis.Options{
	Addr:     os.Getenv(constant.EnvRedisAddress),
	Username: os.Getenv(constant.EnvRedisUsername),
	Password: os.Getenv(constant.EnvRedisUsername),
	DB:       util.StringToInt(os.Getenv(constant.EnvRedisDB), 0),
})

func buildResponseCache(c fiber.Ctx) model.ResponseCache {
	// fmt.Println(c.Response().StatusCode())
	// fmt.Println(string(c.Response().Header.Header()))
	// fmt.Println(string(c.Response().Body()))

	var body model.ResponseBody
	if err := json.Unmarshal(c.Response().Body(), &body); err != nil {
		log.Println("error while build response cache body")
	}

	return model.ResponseCache{
		ResponseStatus:  c.Response().StatusCode(),
		ResponseHeaders: string(c.Response().Header.Header()),
		ResponseBody:    body,
	}
}

func main() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		// Show the latest user ID
		return c.Status(http.StatusOK).JSON(model.ResponseBody{
			Success: true,
			Message: "Success",
			Data: map[string]interface{}{
				"current_user_id": counter.Value(),
			},
		})
	})

	// Middleware
	app.Use(IdempotencyMiddleware())
	app.Post("/user", func(c fiber.Ctx) error {
		// Simulate data storing process
		time.Sleep(3 * time.Second)

		// Simulate increment user ID
		userID := counter.Increment()

		return c.Status(http.StatusCreated).JSON(model.ResponseBody{
			Success: true,
			Message: "success",
			Data: map[string]interface{}{
				"user_id": userID,
			},
		})
	})

	log.Fatal(app.Listen(":3000"))
}

func IdempotencyMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		idempotencyKey := strings.TrimSpace(c.Get(constant.IdempotencyHeaderKey))
		if idempotencyKey == "" {
			log.Println("skip idempotency check...")
			return c.Next()
		}

		fmt.Println("checking idempotency key " + idempotencyKey + "...")

		// 1. Lock idempotency key
		var idempotencyLockKey = fmt.Sprintf("%s-Lock-%s", constant.IdempotencyHeaderKey, idempotencyKey)
		result, err := redisClient.Incr(c.Context(), idempotencyLockKey).Result()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
				Success: false,
				Message: "Something went wrong, please try again later. [Code: 0x00057]",
			})
		}

		fmt.Println("Locking key", idempotencyKey, ":", result)

		if result > 1 {
			return c.Status(http.StatusConflict).JSON(model.ResponseBody{
				Success: false,
				Message: "Duplicate request",
			})
		}

		// 2. Set lock expires in 1 min
		// In case the lock is idle and no action from server to remove the lock, then auto unlock to allow next requests
		if _, err = redisClient.Expire(c.Context(), idempotencyLockKey, 1*time.Minute).Result(); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
				Success: false,
				Message: "Something went wrong, please try again later. [Code: 0x00075]",
			})
		}

		// defer unlock
		defer func() {
			// Unlock idempotency key by deleting lock
			if _, err := redisClient.Del(c.Context(), idempotencyLockKey).Result(); err != nil {
				log.Println("Error while unlock idempotency key: " + idempotencyLockKey)
			}
		}()

		// 3. Check for cached response
		var idempotencyCacheKey = fmt.Sprintf("%s-%s", constant.IdempotencyHeaderKey, idempotencyKey)
		var cacheValue string
		cacheValue, err = redisClient.Get(c.Context(), idempotencyCacheKey).Result()
		if err != nil && err != redis.Nil {
			return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
				Success: false,
				Message: "Something went wrong, please try again later. [Code: 0x00095]",
			})
		}

		// 4a. Return cached response if found
		if cacheValue != "" {
			var responseCache model.ResponseCache
			if err = json.Unmarshal([]byte(cacheValue), &responseCache); err != nil {
				c.Set(constant.IdempotencyRetryHeaderKey, "false")
				return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
					Success: false,
					Message: "Something went wrong, please try again later. [Code: 0x00106]",
				})
			}

			// TODO: compare request body from previous request
			// TODO: return error response can't retry request if the request body is changed

			fmt.Println("returning cached response for idempotency key " + idempotencyKey + "...")

			c.Set(constant.IdempotencyRetryHeaderKey, "true")
			return c.Status(responseCache.ResponseStatus).JSON(responseCache.ResponseBody)
		}

		// 4b. If no cache found, perform request
		c.Next()

		// 5. Cache success response for 24 hours to give the client chances to retry the requests
		responseCache := buildResponseCache(c)
		if responseCache.ResponseStatus == http.StatusOK || responseCache.ResponseStatus == http.StatusCreated {
			cacheValueByte, err := json.Marshal(responseCache)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
					Success: false,
					Message: "Something went wrong, please try again later. [Code: 0x00124]",
				})
			}

			fmt.Println("caching value:", string(cacheValueByte))

			if _, err = redisClient.Set(c.Context(), idempotencyCacheKey, string(cacheValueByte), 24*time.Hour).Result(); err != nil {
				return c.Status(http.StatusInternalServerError).JSON(model.ResponseBody{
					Success: false,
					Message: "Something went wrong, please try again later. [Code: 0x00132]",
				})
			}
		}

		return nil
	}
}
