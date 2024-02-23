package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"bisonai.com/orakl/node/pkg/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Setup(version string) (*fiber.App, error) {
	if version == "" {
		version = "test"
	}

	ctx := context.Background()
	_, err := db.GetPool(ctx)
	if err != nil {
		return nil, errors.New("error getting db pool")
	}

	_, err = db.GetRedisConn(ctx)
	if err != nil {
		return nil, errors.New("error getting redis conn")
	}

	app := fiber.New(fiber.Config{
		AppName:           "Node API " + version,
		EnablePrintRoutes: true,
		ErrorHandler:      CustomErrorHandler,
	})

	app.Use(recover.New(
		recover.Config{
			EnableStackTrace:  true,
			StackTraceHandler: CustomStackTraceHandler,
		},
	))

	app.Use(cors.New())

	return app, nil

}

func CustomErrorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	// Set Content-Type: text/plain; charset=utf-8
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

	// Return status code with error message
	// | ${status} | ${ip} | ${method} | ${path} | ${error}",
	log.Printf("| %d | %s | %s | %s | %s\n", code, c.IP(), c.Method(), c.Path(), err.Error())
	return c.Status(code).SendString(err.Error())
}

func CustomStackTraceHandler(_ *fiber.Ctx, e interface{}) {
	stackTrace := strings.Split(string(debug.Stack()), "\n")
	var failPoint string

	for _, line := range stackTrace {
		if strings.Contains(line, "controller.go") {
			path := strings.Split(strings.TrimSpace(line), " ")[0]
			splitted := strings.Split(path, "/")
			failPoint = splitted[len(splitted)-2] + "/" + splitted[len(splitted)-1]

			break
		}
	}
	log.Printf("| (%s) panic: %v \n", failPoint, e)
	_, _ = os.Stderr.WriteString(fmt.Sprintf("%s\n", debug.Stack())) //nolint:errcheck // This will never fail
}
