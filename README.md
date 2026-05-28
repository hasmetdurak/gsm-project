# GSM (Global Scalable Matrix / Global System Management)

GSM, küresel ölçekte, yüksek performanslı, minimalist ve asenkron iş akışlarına dayalı bir SaaS altyapısı ve sistem yönetimi matrisidir. Yüksek ölçeklenebilirlik, düşük kaynak tüketimi ve Dokploy/VPS dostu mimarisiyle modern bulut operasyonlarını optimize etmek için sıfırdan tasarlanmıştır.

## 🚀 Temel Özellikler

- **Ultra Performanslı Backend:** Go (Golang) tabanlı, asenkron olay döngüleri ve hafif iş akışları ile optimize edilmiş çekirdek motor.
- **Minimalist PWA Arayüzü:** Ağır kütüphaneler (React/Angular) yerine saf, premium ve yüksek hızlı minimalist UI/UX deneyimi.
- **Konteynerizasyon & Kolay Dağıtım:** Docker ve Docker Compose ile paketlenmiş, Dokploy, Coolify veya herhangi bir VPS ortamına tek tıkla dağıtıma hazır mimari.
- **Modüler Altyapı:** Mikroservis geçişlerine ve asenkron kuyruk sistemlerine tam uyumlu, genişletilebilir veri katmanı.

## 📁 Mimari Klasör Yapısı

```text
gsm-project/
├── .prd                 # Ürün Gereksinim Dokümanı
├── plan.md              # Genel Plan ve Yol Haritası
├── docker-compose.yml   # Çoklu Konteyner Orkestrasyonu
├── Dockerfile           # Çok Aşamalı (Multistage) Backend Derleme Dosyası
├── .gitignore           # Git Yoksayma Kuralları
├── main.go              # Go Giriş Noktası & Çekirdek Router
├── core/                # Çekirdek İş Mantığı (Core Logic)
│   └── matrix.go
├── server/              # HTTP API Sunucusu ve Router
│   └── router.go
└── frontend/            # Minimalist PWA Arayüzü
    ├── index.html
    ├── index.css
    └── app.js
```

## 🛠️ Kurulum ve Çalıştırma

### Gereksinimler
- Docker & Docker Compose
- Go (Lokal geliştirme için isteğe bağlı)

### Yerel Geliştirme Ortamı

1. Projeyi klonlayın ve dizine gidin:
   ```bash
   git clone <repo-url>
   cd gsm-project
   ```

2. Konteynerleri ayağa kaldırın:
   ```bash
   docker-compose up --build
   ```

3. Tarayıcınızda erişin:
   - Backend API: `http://localhost:8080`
   - Frontend Minimalist UI: `http://localhost:8080` (veya `frontend/index.html` üzerinden statik servis edilir)

---

## 📄 Lisans
Bu proje MIT Lisansı altında lisanslanmıştır.
