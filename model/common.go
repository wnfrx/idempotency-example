package model

import "github.com/gofiber/fiber/v3"

type ResponseCache struct {
	RequestBody     interface{}  `json:"request_body"`
	ResponseStatus  int          `json:"response_status"`
	ResponseHeaders string       `json:"response_headers"`
	ResponseBody    ResponseBody `json:"response_body"`
}

func (r ResponseCache) buildJSON(c fiber.Ctx) error {
	c.SendStatus(r.ResponseStatus)
	return c.JSON(r.ResponseBody)
}

type ResponseBody struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
