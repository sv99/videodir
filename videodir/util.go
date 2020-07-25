package videodir

import (
	"fmt"
	"github.com/gofiber/fiber"
)

func dumpHeaders(c *fiber.Ctx) {
	auth := c.Get("Authorization")
	fmt.Println(auth)
	c.Next()
}
