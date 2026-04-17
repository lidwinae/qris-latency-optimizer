package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Inisialisasi aplikasi Fiber
	app := fiber.New()

	// Membuat endpoint simulasi Inquiry QRIS sederhana
	app.Get("/inquiry", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
			"message": "Simulasi Inquiry QRIS berhasil. Latency akan diuji di sini.",
		})
	})

	// Menyalakan server di port 3000
	log.Println("Server API berjalan di http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}