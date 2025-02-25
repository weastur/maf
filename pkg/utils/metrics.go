package utils

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	"github.com/gofiber/fiber/v2"
)

func ConfigureMetrics(app *fiber.App) {
	metrics.NewGauge(fmt.Sprintf("maf_version{version=\"%s\"}", AppVersion()), func() float64 {
		return 1
	})

	app.Get("/metrics", func(c *fiber.Ctx) error {
		metrics.WritePrometheus(c.Response().BodyWriter(), true)

		return nil
	})
}
