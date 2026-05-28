package main

import (
	"context"
	"gsm/core"
	"gsm/internal/database"
	"gsm/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	println("--- Starting GSM (Global Scalable Matrix) Bootloader ---")

	// 1. PostgreSQL Veritabanı Bağlantısının Kurulması
	println("[Database] Connecting to PostgreSQL...")
	if err := database.ConnectPostgres(); err != nil {
		println("[Database ERROR] Failed to connect to PostgreSQL:", err.Error())
		println("[Database Warning] Running in offline mode without database features.")
	} else {
		defer database.ClosePostgres()
	}

	// Ana bağlam (context) ve sinyal dinleyici
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Matrix Engine (Asenkron Çekirdek) Başlatılması
	engine := core.NewMatrixEngine()
	engine.Start(ctx)
	println("[Engine] Matrix Core successfully initialized.")

	// 3. HTTP ve API Sunucusunun Yapılandırılması
	addr := ":8080"
	if customPort := os.Getenv("PORT"); customPort != "" {
		addr = ":" + customPort
	}
	
	srv := server.NewServer(addr, engine)

	// Sunucuyu ayrı bir goroutine'de başlatıyoruz
	go func() {
		println("[HTTP] Server is listening on", addr)
		if err := srv.Start(); err != nil {
			println("[HTTP] Error starting server:", err.Error())
		}
	}()

	// 4. Graceful Shutdown (Kibar Kapatma) Yönetimi
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Kapatma sinyalini bekle
	<-stopChan
	println("\n--- Shutting down GSM services gracefully ---")

	// Motoru durdur ve tüm asenkron kanalları kapat
	engine.Stop()
	cancel()

	// Kısa bir bekleme süresi tanıyarak işlemlerin temizlenmesini sağlayalım
	time.Sleep(1 * time.Second)
	println("[System] Shutdown complete. Goodbye, Patron!")
}
