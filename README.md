# FiFriz

AI destekli doküman yönetim sistemi — Go backend, React frontend, PostgreSQL, Redis, MinIO ve Ollama.

## Docker ile çalıştırma (önerilen)

Tüm stack'i tek komutla başlatın:

```bash
docker compose up --build
```

Arka planda:

```bash
docker compose up -d --build
```

Loglar:

```bash
docker compose logs -f backend-api
docker compose logs -f backend-worker
```

Durdur:

```bash
docker compose down
```

Tüm volume'ları sil:

```bash
docker compose down -v
```

### Servisler

| Servis | Container | URL / Port | Notlar |
|--------|-----------|------------|--------|
| Frontend | `fifriz-frontend` | http://localhost:5173 | Nginx, `/api` proxy |
| Backend API | `fifriz-backend-api` | http://localhost:3000 | Health: `/health` |
| Worker | `fifriz-backend-worker` | — | Belge işleme kuyruğu |
| PostgreSQL | `fifriz-postgres` | localhost:5432 | DB: `fifriz`, user/pass: `postgres` |
| Redis | `fifriz-redis` | localhost:6379 | Kuyruk + cache |
| MinIO | `fifriz-minio` | API :9000, Console :9001 | `minioadmin` / `minioadmin`, bucket: `fifiz-documents` |
| Ollama | `fifriz-ollama` | http://localhost:11434 | Embedding + chat modelleri |

### Ortam değişkenleri

- **Docker:** kök dizindeki `.env.docker`
- **Yerel geliştirme:** `backend/.env` (`backend/.env.example` şablonu)

Önemli değişkenler:

| Değişken | Açıklama |
|----------|----------|
| `DATABASE_DSN` | PostgreSQL bağlantısı |
| `REDIS_URL` | Redis bağlantısı |
| `REDIS_QUEUE_KEY` | Worker kuyruk anahtarı (`fifriz:jobs`) |
| `STORAGE_PROVIDER` | `local` veya `minio` |
| `MINIO_*` | MinIO endpoint, bucket, kimlik bilgileri |
| `OLLAMA_BASE_URL` | Ollama API adresi |
| `OLLAMA_EMBEDDING_MODEL` | Varsayılan: `bge-m3` |
| `OLLAMA_CHAT_MODEL` | Varsayılan: `llama3.1:8b` |
| `JWT_SECRET` | Üretimde mutlaka değiştirin |
| `CORS_ALLOWED_ORIGINS` | İzin verilen frontend origin'leri |
| `RUN_IN_PROCESS_WORKER` | Docker'da `false` (ayrı worker container) |

İlk çalıştırmada `ollama-init` şu modelleri indirir:

| Model | Kullanım | Yaklaşık boyut |
|-------|----------|----------------|
| `bge-m3` | Embedding / semantic search | ~1.2 GB |
| `llama3.1:8b` | AI chat | ~4.7 GB |

Modeller `ollama_data` volume'unda saklanır.

Alternatif chat provider: `AI_PROVIDER=deepseek` ve `DEEPSEEK_API_KEY`.

## Yerel geliştirme (Docker dışı)

Altyapı servislerini Docker ile, uygulamayı yerelde çalıştırabilirsiniz:

```bash
docker compose up -d postgres redis minio minio-init ollama ollama-init
```

Backend API:

```bash
cd backend
cp .env.example .env   # gerekirse düzenleyin
go run ./cmd/api
```

Worker (ayrı terminal):

```bash
cd backend
# .env içinde RUN_IN_PROCESS_WORKER=false
go run ./cmd/worker
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Yerel API varsayılan port: `8080`. Frontend `VITE_API_BASE_URL` tanımlı değilse `http://localhost:8080` kullanır.

## Güvenlik notları

- `JWT_SECRET` ve entegrasyon client secret'larını üretimde güçlü değerlerle değiştirin.
- Şifre sıfırlama token'ları hash'li saklanır; düz metin loglanmaz.
- Geliştirme ortamında reset linkleri yalnızca `APP_ENV != production` iken konsola yazılır.
