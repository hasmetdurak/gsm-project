package main

import (
	"context"
	"gsm/core"
	"gsm/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	println("--- Starting GSM (Global Scalable Matrix) Bootloader ---")

	// Ana bağlam (context) ve sinyal dinleyici
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Matrix Engine (Asenkron Çekirdek) Başlatılması
	engine := core.NewMatrixEngine()
	engine.Start(ctx)
	println("[Engine] Matrix Core successfully initialized.")

	// 2. HTTP ve API Sunucusunun Yapılandırılması
	addr := ":8080"
	srv := server.NewServer(addr, engine)

	// Sunucuyu ayrı bir goroutine'de başlatıyoruz
	go func() {
		println("[HTTP] Server is listening on", addr)
		if err := srv.Start(); err != nil {
			println("[HTTP] Error starting server:", err.Error())
		}
	}()

	// 3. Graceful Shutdown (Kibar Kapatma) Yönetimi
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
