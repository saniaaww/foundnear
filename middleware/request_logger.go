
package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		entry := map[string]interface{}{
			"request_id": c.GetRespHeader("X-Request-ID"),
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"duration":   time.Since(start).String(),
		}

		logData, marshalErr := json.Marshal(entry)
		if marshalErr == nil {
			log.Println(string(logData))
		}

		return err
	}
}
