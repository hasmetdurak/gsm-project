package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Event temsil eden yapı
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// MatrixEngine asenkron olayları ve sistem durumunu yöneten çekirdek
type MatrixEngine struct {
	mu         sync.RWMutex
	events     []Event
	listeners  map[string][]chan Event
	isStopped  bool
	cancelFunc context.CancelFunc
}

// NewMatrixEngine yeni bir motor oluşturur
func NewMatrixEngine() *MatrixEngine {
	return &MatrixEngine{
		events:    make([]Event, 0),
		listeners: make(map[string][]chan Event),
	}
}

// Start asenkron iş akışlarını ve sistem döngüsünü başlatır
func (m *MatrixEngine) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	// Örnek arka plan sistem metrikleri üreten asenkron iş akışı
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.Publish(Event{
					ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
					Type:      "SYSTEM_METRIC_TICK",
					Payload:   map[string]any{"cpu_usage": 15.4, "memory_free_gb": 12.8, "active_nodes": 4},
					Timestamp: time.Now(),
				})
			}
		}
	}()
}

// Stop motoru durdurur
func (m *MatrixEngine) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isStopped {
		return
	}
	m.isStopped = true
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

// Publish yeni bir olayı yayınlar ve tüm dinleyicilere asenkron olarak iletir
func (m *MatrixEngine) Publish(event Event) {
	m.mu.Lock()
	m.events = append(m.events, event)
	// Hafızada son 100 olayı tutalım
	if len(m.events) > 100 {
		m.events = m.events[1:]
	}
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Olay tipine göre veya genel dinleyicilere gönderim
	if chs, ok := m.listeners[event.Type]; ok {
		for _, ch := range chs {
			select {
			case ch <- event:
			default:
				// Kanal doluysa tıkanmayı önlemek için geç
			}
		}
	}
	if chs, ok := m.listeners["*"]; ok {
		for _, ch := range chs {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// Subscribe belirli bir olay tipini dinlemek için kanal oluşturur
func (m *MatrixEngine) Subscribe(eventType string) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Event, 10)
	m.listeners[eventType] = append(m.listeners[eventType], ch)
	return ch
}

// GetRecentEvents son olayları listeler
func (m *MatrixEngine) GetRecentEvents() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Slice kopyası oluşturarak thread-safe hale getirelim
	copied := make([]Event, len(m.events))
	copy(copied, m.events)
	return copied
}
