package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gsm/internal/database"
	"gsm/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OauthConfig *oauth2.Config

// InitOauthConfig initializes the Google OAuth config from environment variables
func InitOauthConfig() {
	OauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/documents",
		},
		Endpoint: google.Endpoint,
	}
}

// GenerateStateOauthCookie generates a state token and sets it in cookie for CSRF protection
func GenerateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	
	cookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true, // Canlı ortamda (HTTPS/SaaS) true olmalı
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	return state
}

// HandleGoogleLogin redirects user to Google Consent page
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if OauthConfig == nil {
		InitOauthConfig()
	}
	state := GenerateStateOauthCookie(w)
	// access_type=offline parametresi refresh_token alabilmek için kritik önem taşır!
	url := OauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes Google OAuth callback
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if OauthConfig == nil {
		InitOauthConfig()
	}

	// CSRF Koruma Kontrolü
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		http.Error(w, "Invalid OAuth state (CSRF Protection)", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := OauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Google API'den kullanıcı e-postasını çekelim
	client := OauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Kullanıcıyı veritabanına kaydet / güncelle
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userID string
	var subscriptionType string

	// Google'dan her zaman refresh_token dönmeyebilir (eğer kullanıcı daha önce izin verdiyse).
	// Bu nedenle sadece döndüğünde (boş değilse) refresh_token'ı güncelleyelim.
	query := `
		INSERT INTO users (email, google_id, access_token, refresh_token, token_expiry, subscription_type)
		VALUES ($1, $2, $3, $4, $5, 'free')
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			access_token = EXCLUDED.access_token,
			refresh_token = CASE 
				WHEN EXCLUDED.refresh_token <> '' THEN EXCLUDED.refresh_token 
				ELSE users.refresh_token 
			END,
			token_expiry = EXCLUDED.token_expiry,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, subscription_type;
	`
	
	refreshToken := ""
	if token.RefreshToken != "" {
		refreshToken = token.RefreshToken
	}

	err = database.Pool.QueryRow(ctx, query, 
		googleUser.Email, 
		googleUser.ID, 
		token.AccessToken, 
		refreshToken, 
		token.Expiry,
	).Scan(&userID, &subscriptionType)

	if err != nil {
		http.Error(w, "Database upsert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcıya platform erişimi için JWT üretelim (SaaS güvenliği)
	jwtToken, err := GenerateJWT(userID, googleUser.Email, subscriptionType)
	if err != nil {
		http.Error(w, "Failed to generate session JWT", http.StatusInternalServerError)
		return
	}

	// JWT'yi güvenli HttpOnly cookie olarak set edelim
	cookie := &http.Cookie{
		Name:     "gsm_session",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	// Başarılı girişten sonra frontend paneline yönlendir
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GenerateJWT generates session token for UI / API communication
func GenerateJWT(userID, email, subType string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "gsm_default_secret_key_change_in_production"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":           userID,
		"email":             email,
		"subscription_type": subType,
		"exp":               time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// RefreshTokenIfNeeded checks if Google access_token has expired.
// If expired, it utilizes refresh_token to acquire a new access_token,
// updates the PostgreSQL database, and returns the active access_token.
func RefreshTokenIfNeeded(ctx context.Context, user *models.User) (string, error) {
	// Token'ın geçerlilik süresinin bitmesine 2 dakikadan az kaldıysa veya bittiyse yenileyelim
	if time.Until(user.TokenExpiry) > 2*time.Minute {
		return user.AccessToken, nil
	}

	if user.RefreshToken == "" {
		return "", fmt.Errorf("refresh token is empty, re-authentication required")
	}

	if OauthConfig == nil {
		InitOauthConfig()
	}

	// OAuth2 kütüphanesi kullanarak token yenileme isteği gönderelim
	tokenSource := OauthConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: user.RefreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to refresh google token: %w", err)
	}

	// Veritabanını yeni token ve expiry bilgileriyle güncelleyelim
	updateQuery := `
		UPDATE users 
		SET access_token = $1, 
		    token_expiry = $2, 
		    updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3;
	`
	_, err = database.Pool.Exec(ctx, updateQuery, newToken.AccessToken, newToken.Expiry, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to update refreshed tokens in database: %w", err)
	}

	println("[OAuth] Access Token successfully auto-rotated for user:", user.Email)
	return newToken.AccessToken, nil
}

// GetUserFromSession decodes session JWT and returns user details from DB
func GetUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("gsm_session")
	if err != nil {
		return nil, fmt.Errorf("session cookie not found")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "gsm_default_secret_key_change_in_production"
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in token claims")
	}

	// Veritabanından en güncel kullanıcı bilgilerini çekelim
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User
	query := `SELECT id, email, google_id, access_token, refresh_token, token_expiry, subscription_type, created_at, updated_at FROM users WHERE id = $1`
	err = database.Pool.QueryRow(ctx, query, userID).Scan(
		&u.ID, &u.Email, &u.GoogleID, &u.AccessToken, &u.RefreshToken, &u.TokenExpiry, &u.SubscriptionType, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found in database")
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	return &u, nil
}
