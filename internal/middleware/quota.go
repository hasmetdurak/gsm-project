package middleware

import (
	"context"
	"encoding/json"
	"gsm/internal/auth"
	"gsm/internal/database"
	"gsm/internal/models"
	"net/http"
	"time"
)

type QuotaErrorResponse struct {
	Error string `json:"error"`
}

// UserContextKey represents key for storing user in context
type UserContextKey string
const UserKey UserContextKey = "active_user"

// QuotaMiddleware tracks user usage quota and restricts free users to 100 daily actions
func QuotaMiddleware(actionType string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Session'dan kullanıcıyı doğrula
		user, err := auth.GetUserFromSession(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized session. Please login via Google OAuth."})
			return
		}

		// 2. Kota Kontrolü (Premium ise direkt geçiş, Free ise 100 işlem sınırı)
		if user.SubscriptionType != "premium" {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Son 24 saat içindeki başarılı kullanım kaydı sayısını çekelim
			query := `
				SELECT COUNT(*) 
				FROM usage_logs 
				WHERE user_id = $1 
				  AND timestamp >= $2;
			`
			
			// 24 saat öncesini hesapla
			twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
			
			var count int
			err = database.Pool.QueryRow(ctx, query, user.ID, twentyFourHoursAgo).Scan(&count)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify usage quota: " + err.Error()})
				return
			}

			// Günlük 100 işlem limiti aşıldıysa HTTP 423 Locked dönelim!
			if count >= 100 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusLocked) // HTTP 423
				json.NewEncoder(w).Encode(QuotaErrorResponse{
					Error: "Daily quota exceeded. Upgrade to Premium for 9.99$/month for unlimited access.",
				})
				return
			}
		}

		// 3. Kullanıcı bilgilerini context'e ekleyerek sonraki handler'ların kullanmasını sağlayalım
		ctx := context.WithValue(r.Context(), UserKey, user)
		r = r.WithContext(ctx)

		// 4. İsteği çalıştır
		next.ServeHTTP(w, r)

		// 5. İşlem başarılı bittiğinde asenkron (goroutine) olarak kullanım logu ekleyelim (Yüksek Performans için!)
		go LogUsage(user.ID, actionType)
	}
}

// LogUsage registers a new action log in database asynchronously
func LogUsage(userID, actionType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO usage_logs (user_id, action_type, timestamp)
		VALUES ($1, $2, CURRENT_TIMESTAMP);
	`
	_, err := database.Pool.Exec(ctx, query, userID, actionType)
	if err != nil {
		println("[Quota Error] Failed to log usage in database:", err.Error())
	}
}

// GetUserFromContext retrieves authenticated user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserKey).(*models.User)
	return user, ok
}
