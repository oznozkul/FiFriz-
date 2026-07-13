package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	CORS      CORSConfig
	Storage   StorageConfig
	Embedding EmbeddingConfig
	AI        AIConfig
	JWT       JWTConfig
	Public    PublicConfig
	PasswordReset PasswordResetConfig
	Google        GoogleOAuthConfig
	GoogleDrive   GoogleDriveOAuthConfig
	Microsoft     MicrosoftOAuthConfig
	MicrosoftOneDrive MicrosoftOAuthConfig
	Integrations IntegrationsConfig
	Documents    DocumentsConfig
}

type PublicConfig struct {
	AppURL string
	APIURL string
}

type PasswordResetConfig struct {
	FrontendURL string
	TokenTTL    time.Duration
}

type AppConfig struct {
	Env string
}

type HTTPConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ReadBufferSize  int
	ShutdownTimeout time.Duration
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type DatabaseConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	QueueKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type StorageConfig struct {
	Provider      string
	LocalBasePath string
	ProcessingDir string
	MinIO         MinIOStorageConfig
}

type MinIOStorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

type EmbeddingConfig struct {
	Provider string
	Ollama   OllamaEmbeddingConfig
}

type OllamaEmbeddingConfig struct {
	BaseURL string
	Model   string
}

type AIConfig struct {
	Provider string
	Ollama   OllamaChatConfig
	DeepSeek DeepSeekConfig
}

type OllamaChatConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type JWTConfig struct {
	Secret   string
	Issuer   string
	Duration time.Duration
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type MicrosoftOAuthConfig struct {
	ClientID     string
	ClientSecret string
	Tenant       string
	RedirectURL  string
	Scopes       []string
}

type GoogleDriveOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type IntegrationsConfig struct {
	SyncInterval time.Duration
}

type DocumentsConfig struct {
	MaxFileSizeBytes      int64
	AllowedMIMETypes      []string
	StorageDownloadRetries int
	StorageDownloadTimeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("APP_HOST", "0.0.0.0"),
			Port:            getEnv("APP_PORT", "8080"),
			ReadTimeout:     getDuration("APP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDuration("APP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getDuration("APP_IDLE_TIMEOUT", 60*time.Second),
			ReadBufferSize:  getInt("APP_READ_BUFFER_SIZE", 64*1024),
			ShutdownTimeout: getDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             buildDatabaseDSN(),
			MaxIdleConns:    getInt("DATABASE_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getInt("DATABASE_MAX_OPEN_CONNS", 50),
			ConnMaxLifetime: getDuration("DATABASE_CONN_MAX_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			QueueKey: getEnv("REDIS_QUEUE_KEY", "fileai:jobs"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		},
		Storage: StorageConfig{
			Provider:      getEnv("STORAGE_PROVIDER", "local"),
			LocalBasePath: firstNonEmpty(getEnv("LOCAL_STORAGE_PATH", ""), getEnv("DOCUMENT_STORAGE_DIR", "storage/documents")),
			ProcessingDir: getEnv("STORAGE_PROCESSING_DIR", "tmp/processing"),
			MinIO: MinIOStorageConfig{
				Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
				AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
				SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
				Bucket:    getEnv("MINIO_BUCKET", "fifiz-documents"),
				UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
				Region:    getEnv("MINIO_REGION", "us-east-1"),
			},
		},
		Embedding: EmbeddingConfig{
			Provider: getEnv("EMBEDDING_PROVIDER", "ollama"),
			Ollama: OllamaEmbeddingConfig{
				BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
				Model:   getEnv("OLLAMA_EMBEDDING_MODEL", "bge-m3"),
			},
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "ollama"),
			Ollama: OllamaChatConfig{
				BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
				Model:   getEnv("OLLAMA_CHAT_MODEL", "llama3.1:8b"),
				Timeout: getDuration("OLLAMA_CHAT_TIMEOUT", 5*time.Minute),
			},
			DeepSeek: DeepSeekConfig{
				APIKey:  getEnv("DEEPSEEK_API_KEY", ""),
				BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
				Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
			},
		},
		JWT: JWTConfig{
			Secret:   getEnv("JWT_SECRET", "dev-change-me-fifriz-secret"),
			Issuer:   getEnv("JWT_ISSUER", "fifriz"),
			Duration: getDuration("JWT_DURATION", 24*time.Hour),
		},
		Public: PublicConfig{
			AppURL: getEnv("PUBLIC_APP_URL", "http://localhost:5173"),
			APIURL: getEnv("PUBLIC_API_URL", "http://localhost:8080"),
		},
		PasswordReset: PasswordResetConfig{
			FrontendURL: firstNonEmpty(getEnv("APP_FRONTEND_URL", ""), getEnv("PUBLIC_APP_URL", "http://localhost:5173")),
			TokenTTL:    time.Duration(getInt("PASSWORD_RESET_TOKEN_TTL_MINUTES", 30)) * time.Minute,
		},
		Google: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/integrations/gmail/callback"),
			Scopes:       getStringSlice("GOOGLE_GMAIL_SCOPES", []string{"https://www.googleapis.com/auth/gmail.readonly"}),
		},
		GoogleDrive: GoogleDriveOAuthConfig{
			ClientID:     getEnv("GOOGLE_DRIVE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_DRIVE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_DRIVE_REDIRECT_URL", "http://localhost:3000/api/v1/integrations/google-drive/callback"),
			Scopes:       getStringSlice("GOOGLE_DRIVE_SCOPES", []string{"https://www.googleapis.com/auth/drive.readonly"}),
		},
		Microsoft: MicrosoftOAuthConfig{
			ClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
			ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),
			Tenant:       getEnv("MICROSOFT_TENANT", "common"),
			RedirectURL:  getEnv("MICROSOFT_REDIRECT_URL", "http://localhost:8080/api/v1/integrations/outlook/callback"),
			Scopes:       parseScopeList(getEnv("MICROSOFT_SCOPES", "offline_access User.Read Mail.Read")),
		},
		MicrosoftOneDrive: MicrosoftOAuthConfig{
			ClientID:     getEnv("MICROSOFT_ONEDRIVE_CLIENT_ID", ""),
			ClientSecret: getEnv("MICROSOFT_ONEDRIVE_CLIENT_SECRET", ""),
			Tenant:       getEnv("MICROSOFT_ONEDRIVE_TENANT", "common"),
			RedirectURL:  getEnv("MICROSOFT_ONEDRIVE_REDIRECT_URL", "http://localhost:3000/api/v1/integrations/onedrive/callback"),
			Scopes:       parseScopeList(getEnv("MICROSOFT_ONEDRIVE_SCOPES", "offline_access User.Read Files.Read.All")),
		},
		Integrations: IntegrationsConfig{
			SyncInterval: getDuration("INTEGRATION_SYNC_INTERVAL", 15*time.Minute),
		},
		Documents: DocumentsConfig{
			MaxFileSizeBytes:       int64(getInt("DOCUMENT_MAX_FILE_SIZE_BYTES", 100*1024*1024)),
			AllowedMIMETypes:       getStringSlice("DOCUMENT_ALLOWED_MIME_TYPES", defaultDocumentMIMEAllowlist()),
			StorageDownloadRetries: getInt("STORAGE_DOWNLOAD_RETRIES", 3),
			StorageDownloadTimeout: getDuration("STORAGE_DOWNLOAD_TIMEOUT", 30*time.Second),
		},
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database connection is required: set DATABASE_DSN or DATABASE_HOST/DATABASE_USER/DATABASE_PASSWORD/DATABASE_NAME")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// buildDatabaseDSN prefers DATABASE_DSN when set; otherwise builds from parts.
func buildDatabaseDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_DSN")); dsn != "" {
		return dsn
	}

	host := getEnv("DATABASE_HOST", "")
	user := getEnv("DATABASE_USER", "")
	password := getEnv("DATABASE_PASSWORD", "")
	name := getEnv("DATABASE_NAME", "")
	if host == "" || user == "" || name == "" {
		// Local development fallback when neither DSN nor parts are configured.
		return "host=localhost user=postgres password=postgres dbname=fifriz port=5432 sslmode=disable TimeZone=UTC"
	}

	port := getEnv("DATABASE_PORT", "5432")
	sslMode := getEnv("DATABASE_SSLMODE", "disable")
	timezone := getEnv("DATABASE_TIMEZONE", "UTC")

	parts := []string{
		"host=" + host,
		"user=" + user,
		"password=" + password,
		"dbname=" + name,
		"port=" + port,
		"sslmode=" + sslMode,
		"TimeZone=" + timezone,
	}
	return strings.Join(parts, " ")
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getStringSlice(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return fallback
	}
	return result
}

func parseScopeList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.Contains(value, ",") {
		return getStringSliceFromValue(value)
	}
	return strings.Fields(value)
}

func getStringSliceFromValue(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func defaultDocumentMIMEAllowlist() []string {
	return []string{
		"application/pdf",
		"image/png",
		"image/jpeg",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
