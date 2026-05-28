-- UUID uzantısını aktifleştirelim
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Kullanıcılar tablosu (Google OAuth bilgilerini ve üyelik tipini tutar)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    subscription_type VARCHAR(50) DEFAULT 'free', -- 'free' veya 'premium'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Günlük kullanım logları tablosu (Kota kontrolü için her başarılı MCP isteğini loglar)
CREATE TABLE IF NOT EXISTS usage_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL, -- 'sheets_write', 'gmail_send', 'docs_create' vb.
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Kota sorgularını jet hızına indirmek için index'leri oluşturalım
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_date ON usage_logs(user_id, timestamp);
