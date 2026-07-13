# Project Context

Bu proje AI destekli doküman yönetim sistemi olacak.

## Amaç

Kullanıcılar PDF, PNG, JPG gibi belgelerini yükleyebilecek. Sistem belgelerin içindeki yazıları çıkaracak, veritabanına kaydedecek, klasik arama ve semantic search yapacak. Sonrasında kullanıcı belgeleriyle AI chat üzerinden konuşabilecek.

## Kullanım Modları

- Kişisel kullanıcı
- Kurumsal kullanıcı

Kişisel kullanıcı kira sözleşmesi, fatura, garanti belgesi, hukuki evrak gibi belgeleri yükleyebilir.

Kurumsal kullanıcı şirket, departman, çalışan, İK evrakları, sözleşmeler ve faturaları yönetebilir.

## Stack

Backend:

- Go
- GORM
- PostgreSQL
- pgvector
- Redis
- MinIO

Frontend:

- React
- Vite
- Tailwind CSS

Mobile future:

- React Native / Expo

AI:

- Ollama Embeddings (`bge-m3`)
- Ollama Chat (`llama3.1:8b`)
- DeepSeek Chat (alternatif provider)
- RAG mimarisi

## Temel Akış

1. Kullanıcı belge yükler.
2. Dosya MinIO’ya kaydedilir.
3. PostgreSQL’de documents kaydı açılır.
4. Worker belgeyi işler.
5. PDF içinden text çıkarılır.
6. Text çıkmazsa OCR yapılır.
7. Metin temizlenir.
8. Metin chunklara bölünür.
9. Her chunk için embedding oluşturulur.
10. Embedding pgvector’a kaydedilir.
11. Kullanıcı klasik arama veya semantic search yapabilir.
12. AI chat, ilgili chunkları bulup cevap üretir.

## DB Mantığı

Ana tablolar:

- users
- documents
- document_texts
- document_chunks
- folders
- tags
- chat_sessions
- chat_messages

Vektörel arama için:

document_chunks tablosunda embedding VECTOR(1536) tutulacak.

## İlk MVP

Önce web yapılacak.

Ekranlar:

- Dashboard
- Belgelerim
- Belge Detay
- AI Asistan
- Soru & Cevap Geçmişi
- Ayarlar

İlk amaç:

- kullanıcı kayıt/giriş
- belge yükleme
- belge listeleme
- belge arama
- semantic search
- AI asistan ekranı

## Frontend Tasarım

Modern, sade, beyaz ağırlıklı, mor/violet accent kullanılan dashboard tasarımı.

**Frontend geliştirmelerinde `DESIGN_SYSTEM.md` kuralları takip edilmelidir.**

Ana layout:

- Sol sidebar
- Üst search bar
- Belge yükle butonu
- Dashboard kartları
- Belge tablosu
- AI chat paneli

## Kodlama Prensipleri

Go tarafında idiomatic Go kullanılacak.

- context.Context kullanılmalı
- explicit error handling
- GORM kullanılacak
- gereksiz abstraction yapılmayacak
- modular monolith yapı kurulacak
- tenant/user isolation unutulmayacak

Frontend tarafında:

- React + TypeScript
- Tailwind CSS
- component bazlı yapı
- mock data ile başlanacak
- sonra Go API’ye bağlanacak

### Bildirim ve yükleme (UX)

Uygulama genelinde tutarlı geri bildirim için ortak bileşenler kullanılır:

- `frontend/src/services/toast.ts` — `toast.success`, `toast.error`, `toast.info`, `toast.loading`, `toast.promise`
- `frontend/src/components/common/ToastProvider.tsx` — `main.tsx` içinde mount edilir; sağ alt köşede toast gösterir
- `frontend/src/components/common/Spinner.tsx` — `sm` / `md` / `lg` boyutlarında dönen yükleme göstergesi
- `frontend/src/components/common/LoadingState.tsx` — `inline`, `panel`, `fullscreen` varyantları

**Ne zaman toast?**

- API hataları (`catch` blokları)
- Başarılı mutasyonlar (oluştur, güncelle, sil, yükle, favori)
- Bilgilendirme (yenileme, arka plan işlemi tamamlandı)

**Ne zaman spinner?**

- Sayfa veya panel ilk yüklemesi → `LoadingState` (`panel` veya `fullscreen`)
- Buton/form gönderimi → `Spinner` buton içinde
- Küçük alanlar (arama, dropdown, favori yıldızı) → `Spinner size="sm"`

**Ne zaman inline hata bandı kullanılmaz?**

- Genel sayfa hataları artık toast ile gösterilir; kırmızı banner yerine `EmptyState` veya toast tercih edilir.
- Form/modal içi geçici hatalar da mümkün olduğunda toast'a taşınır.

Mesajlar Türkçe ve kısa olmalıdır (ör. "Belge yüklendi.", "Klasör silinemedi.").

### Dosya turu ikonlari

Belgelerde generic `FileText` ikonlari yerine gercek dosya turu ikonlari kullanilir:

- `frontend/src/components/common/FileTypeIcon.tsx` — uzantiya gore ikon (PDF, Word, Excel, gorsel, diger)
- `react-icons/fa6` — `FaFilePdf`, `FaFileWord`, `FaFileExcel`, `FaFileImage`
- Kullanim: `<FileTypeIcon extension={document.file_extension} />`
- Dosya adindan uzanti: `extensionFromFileName(file.name)`

## Architecture Principles

Bundan sonraki tum gelistirmeler asagidaki prensiplere uymalidir.

### Multi Tenant First

Her servis tenant-aware olmalidir. Hicbir veri `tenant_id` filtresi olmadan okunmamalıdır.

### Event Driven Architecture

CRUD islemleri icinde dogrudan `cache.Delete`, `notification.Create`, `audit.Create`, `search.Refresh` cagirilmaz. CRUD yalnizca **domain event** publish eder. Yan etkiler `internal/listeners` uzerinden calisir.

Paketler:

- `internal/events` — EventBus, Publisher, Dispatcher
- `internal/events/events` (`domain`) — DocumentCreated, FolderUpdated, vb.
- `internal/listeners` — AuditListener, NotificationListener, CacheListener, CompressionListener

### Transaction After Commit

Domain event yalnizca transaction commit basarili olursa publish edilir (`internal/tx`). Rollback olursa cache temizlenmez, bildirim/audit olusmaz.

```go
tx.WithTransaction(ctx, db, func(ctx context.Context) error {
    // CRUD
    publisher.PublishAfterCommit(ctx, domain.NewDocumentCreated(...))
    return nil
})
```

### Queue First

Uzun suren islemler HTTP request icinde yapilmaz: upload → queue → worker → processing. Ornek job tipleri: `process_document`, `sync_integration`, `compress_document`.

### Plugin Based Architecture

Yeni entegrasyon provider eklemek icin mevcut kod degistirilmez. `internal/integrations/registry` uzerinden plugin register edilir.

### Redis Cache Strategy

Sik kullanilan read endpointleri `internal/cache` uzerinden `GetOrSet` kullanir. Cache key standardi:

`tenant:{tenantId}:module:{module}:resource:{resource}`

TTL ornekleri: Dashboard 60s, Folder Tree 5dk, Document Detail/Workspace 2dk, Tags 5dk, Notifications 30s, Navigation Pins/Recent 5dk.

### Workspace Navigation

Uygulama navigasyonu route metadata uzerinden uretilir; breadcrumb hardcoded degildir.

- `frontend/src/data/routeMetadata.ts` — statik route etiketleri ve bolumler
- `frontend/src/components/common/Breadcrumb.tsx` — reusable breadcrumb (overflow `...` destegi)
- `frontend/src/context/BreadcrumbContext.tsx` — dinamik breadcrumb (belge klasor hiyerarsisi vb.)
- Topbar: solda breadcrumb, ortada global search (Ctrl+K), sagda bildirimler ve kullanici menusu
- Sidebar: aktif route vurgusu, collapse destegi, alt kisimda Hizli Erisim (pinned + recent)

### Document Workspace

`/documents/:id` route'u artik tek bir calisma alani (Document Workspace) sunar. Preview bozulmaz; mevcut `DocumentFilePreview` bileseni workspace icinde kullanilir.

**Layout:** Toolbar | Preview Area | Right Information Panel (accordion)

- Toolbar: belge adi, breadcrumb, ac/indir/paylas/favori, workflow/versiyon/aktivite/AI (placeholder) kisayollari, tam ekran (F / Esc), panel collapse
- Preview: PDF, DOCX, XLSX, TXT, image — mevcut preview davranisi korunur
- Sag panel: Genel Bilgiler (checksum, depolama, MIME, kodlama, sikistirma, lifecycle badge), Belge Sagligi, Etiketler, Klasor, Paylasimlar, Onay Sureci, Yorumlar, Versiyonlar, Onemli Tarihler, Aktivite, Ilgili Belgeler, AI (placeholder)
- Responsive: mobilde sag panel drawer olarak acilir
- Klavye: `F` tam ekran, `Esc` cikis

**Frontend modulleri:** `frontend/src/components/documentWorkspace/`

- `DocumentWorkspace.tsx` — layout orchestrator
- `DocumentFilePreview.tsx` — preview + arama vurgulama (eski DocumentDetailPage'den ayrildi)
- `DocumentWorkspaceToolbar.tsx`, `DocumentWorkspacePanel.tsx` — toolbar ve accordion panel

Mevcut panel bilesenleri (`DocumentTagsPanel`, `DocumentCommentsPanel`, vb.) yeniden kullanilir; yeni business logic eklenmez.

**Backend cache:** `GET /api/v1/documents/:id/workspace` — belge detayi Redis'te 2 dakika TTL (`DocumentWorkspaceKey`). Her acilista `document.opened` eventi publish edilir (recent sistemi dinler).

### Folder Management

Klasor modulu enterprise seviyesinde: sinirsiz derinlikte agac (`parent_id`), lifecycle, kilit, metadata, cascade soft delete, izin devralma ve gelismis arama.

**API (auth):**

- `GET /api/v1/folders` — nested tree
- `GET /api/v1/folders/search?q=` — isim, aciklama, AI siniflandirma, belge sayisi
- `POST /api/v1/folders` — olustur (`parentId`, `icon`, `color`)
- `PUT /api/v1/folders/:id` — guncelle (ad, aciklama, status, icon, color)
- `DELETE /api/v1/folders/:id` — cascade soft delete (belgeler silinmez)
- `POST /api/v1/folders/:id/move` — tasi (cycle/self/name/lock kontrolu)
- `POST /api/v1/folders/:id/lock|unlock` — owner/admin
- `POST /api/v1/folders/:id/restore` — soft delete geri al
- `GET /api/v1/folders/:id/metadata` — belge/alt klasor sayisi, boyut, AI indexed
- `GET /api/v1/folders/:id/breadcrumb` — tree tabanli yol
- `GET|PUT /api/v1/folders/:id/permissions` — izin devralma ve override

### Folder Tree

Adjacency list (`folders.parent_id`). `internal/folders/tree.go`: cycle detection, descendant collection, breadcrumb builder.

Tasima kurallari: kendi altina tasinamaz, ayni parent altinda ayni isim olamaz, tenant disina cikamaz, kilitli klasor tasinamaz.

### Folder Lifecycle

Status: `active`, `archived`, `hidden`. `is_locked` ayri flag; UI `lifecycleStatus` badge (`locked`, `deleted` dahil).

Silme: klasor + alt klasorler soft delete; `documents.folder_id` korunur.

### Permission Inheritance

`inherit_permissions` (default true). `folder_permissions` tablosu: tenant, folder, user, permission (`view`/`edit`/`manage`). Inheritance kirilabilir; kilitli klasorde izin degisikligi engellenir.

Audit: `folder_created`, `folder_renamed`, `folder_deleted`, `folder_moved`, `folder_locked`, `folder_unlocked`, `folder_permission_changed`.

Frontend: `FoldersPage` — tree (expand/collapse, drag-drop, sag tik, renk/ikon, sabitleme, son acilan), `FolderMetadataPanel`, `FolderMoveModal`.

### Document Lifecycle

Belge durumu `documents.status` kolonunda tutulur. Enterprise durumlar:

- `uploading` — dosya yukleniyor veya is kuyruguna aliniyor
- `processing` — worker metin/OCR/embedding isliyor
- `ready` — kullanima hazir (eski `processed` degeri okunurken `ready` olarak normalize edilir)
- `failed` — isleme hatasi
- `archived` — arsivlenmis
- `deleted` — soft delete / cop kutusu
- `locked` — UI badge icin efektif durum (`is_locked=true` iken status ne olursa olsun)
- `expired` — `expires_at` gecmis
- `retention_locked` — `retention_locked_at` dolu

API yaniti `lifecycleStatus` alanini dondurur. Frontend `StatusBadge` bu degeri gosterir.

### Document Lock

`documents.is_locked`, `locked_at`, `locked_by_user_id` kolonlari.

Kilitli belgede engellenen islemler: yeniden adlandirma, silme, uzerine yazma, yeni versiyon yukleme.

- `POST /api/v1/documents/:id/lock`
- `POST /api/v1/documents/:id/unlock` — owner veya admin

### Document Health

`GET /api/v1/documents/:id/health` — belge saglik paneli verisi:

- `previewGenerated`, `ocrGenerated`, `embeddingGenerated`, `summaryGenerated`, `metadataGenerated`, `aiCompleted`
- `versionCount`, `storageStatus`

Workspace sag panelde **Belge Sagligi** accordion bolumu bu endpoint'i kullanir.

### Document Validation

Upload ve versiyon yuklemede:

- **SHA-256 checksum** hesaplanir (`sha256_checksum`); ayni tenant'ta eslesen belge varsa yalnizca uyari (`warnings`, `duplicateOf`) — yukleme engellenmez
- **Magic bytes** ile MIME dogrulama (`internal/documents/validation/mime.go`); uzanti ile icerik uyumsuzsa red
- **Boyut ve MIME allowlist** — global `DOCUMENT_MAX_FILE_SIZE_BYTES`, `DOCUMENT_ALLOWED_MIME_TYPES`; tenant override `tenant_document_policies` tablosu uzerinden `PolicyResolver` ile birlestirilir
- **VirusScanner** arayuzu (`internal/documents/virusscan`); varsayilan `NoOpScanner`, ileride ClamAV takilabilir

Indirme dayanikliligi: `storage.DownloadWithRetry` — `STORAGE_DOWNLOAD_RETRIES`, `STORAGE_DOWNLOAD_TIMEOUT`.

Dosya erisim audit: `document_previewed`, `document_downloaded` (auth), `share_link_downloaded` (public share indirme).

### Document Graph

Belgeler yonlu kenarlarla (`DocumentRelationship`) birbirine baglanir. Hem personal hem organization tenantlarda tenant izolasyonu korunur.

**Model:** `document_relationships`

- `tenant_id`, `source_document_id`, `target_document_id`, `relationship_type`, `created_by_user_id`, `created_at`
- Unique: `(tenant_id, source_document_id, target_document_id, relationship_type)`
- Belge kendisiyle iliskilendirilemez; ayni kenar iki kez olusturulamaz

**Relationship types:** `related`, `attachment`, `parent`, `child`, `replacement`, `reference`, `invoice_of`, `contract_of`, `identity_of`, `custom`

**API (AI / Smart Search uyumlu):**

- `GET /api/v1/documents/:id/relationships` — `items` (edge listesi + `relatedDocument` node) ve `summary` (`byType`, incoming/outgoing sayilari)
- `POST /api/v1/documents/:id/relationships` — body: `targetDocumentId`, `relationshipType` (kaynak = path'teki belge)
- `DELETE /api/v1/document-relationships/:id`

**Izinler:** Listeleme = belge goruntuleme; olusturma = kaynak belgede duzenleme + hedef belgede goruntuleme; silme = olusturan veya kaynak belgede duzenleme veya owner/admin.

**Workspace:** Sag panelde `Iliskili Belgeler` — belge, iliski turu, ac, kaldir; belge arama ile hedef secimi (`/api/v1/documents/search`).

### Recent Items

Son acilan ogeler event-driven calisir; CRUD icinde dogrudan recent kaydi olusturulmaz.

- Domain eventler: `document.opened`, `folder.opened`, `smart_collection.opened`, `integration.opened`, `settings.page.opened`
- `internal/listeners/navigation_listener.go` — RecentListener Redis'e yazar (son 20 kayit, tenant+user bazli)
- `GET /api/v1/recent` — gruplanmis son acilanlar
- `POST /api/v1/navigation/visited` — ayar/koleksiyon ziyaretlerini event olarak publish eder

### Pinned Items

- Model: `pinned_items` (tenant_id, user_id, entity_type, entity_id, label, path)
- `GET /api/v1/pins`, `POST /api/v1/pins`, `DELETE /api/v1/pins/:id`
- Eventler: `navigation.pinned.created`, `navigation.pinned.deleted` — cache invalidation icin

Cache invalidation CRUD icinde yapilmaz; **CacheListener** domain eventlerini dinler.

### Storage Compression Strategy

`internal/storage/compression` — txt, csv, json, xml, html, eml dosyalari arka planda gzip ile sikistirilir. PDF, DOCX, XLSX, PNG, JPG default olarak sikistirilmez. `compress_document` job ile calisir; upload akisi bozulmaz.

### Storage Strategy

- Local storage sadece development/MVP icin.
- Production storage MinIO/S3 compatible provider uzerinden calisir.
- `internal/storage` paketindeki `StorageProvider` interface kullanilmalidir (`LocalStorageProvider`, `MinIOStorageProvider`).
- Business logic local path bilmemeli; dosya erisimi `StorageProvider` uzerinden yapilir.
- Tenant-aware object keys kullanilir:
  - `tenants/{tenantId}/documents/{documentId}/original/{storedFileName}`
  - `tenants/{tenantId}/documents/{documentId}/versions/{versionNumber}/{storedFileName}`
  - `tenants/{tenantId}/integrations/{integrationAccountId}/documents/{documentId}/{storedFileName}`
- Worker dosyayi provider uzerinden indirip `tmp/processing` altinda gecici dosyaya yazarak text extraction yapar.
- Preview/download permission kontrolunden gecmeden dosya acilmaz; public share linklerde presigned URL dogrudan disari verilmez, backend stream eder.
- Mevcut `storage_provider=local` (veya bos) belgeler LocalStorageProvider ile calismaya devam eder.
- Yeni yuklemeler `STORAGE_PROVIDER` env degerine gore MinIO veya local'e gider.
- Opsiyonel migration: `POST /api/v1/admin/storage/migrate-local-to-minio` (owner/admin veya development).

### Extension System

`internal/extensions` — gelecekte AI, Webhook, Slack, Teams, ERP gibi sistemler yalnizca Extension olarak eklenir. `Extension` interface: `Name()`, `Enabled()`, `Handle(event)`, `Register()`.

### AI Provider Strategy

FiFriz AI katmani provider-based calisir (`internal/ai`, `internal/embedding`).

| Kullanim | Provider | Model | Env |
|----------|----------|-------|-----|
| Embedding / semantic search | Ollama | `bge-m3` | `EMBEDDING_PROVIDER=ollama`, `OLLAMA_EMBEDDING_MODEL=bge-m3` |
| Chat / answer generation | Ollama | `llama3.1:8b` | `AI_PROVIDER=ollama`, `OLLAMA_CHAT_MODEL=llama3.1:8b` |

Ortak Ollama ayarlari:

- `OLLAMA_BASE_URL` — varsayilan `http://localhost:11434` (Docker: `http://ollama:11434`)

Chat provider:

- `internal/ai.ChatProvider` interface
- `OllamaChatProvider` — Ollama `/api/chat` endpoint
- `DeepSeekChatProvider` — alternatif; silinmez, `AI_PROVIDER=deepseek` ile secilebilir

Docker `ollama-init` servisi ilk calistirmada `bge-m3` ve `llama3.1:8b` modellerini indirir.

### AI Pipeline Architecture

Belge isleme surecindeki tum AI islemleri `internal/ai/pipeline` paketi uzerinden calisir. Her yeni AI ozelligi **Pipeline Step** olarak eklenmelidir; hardcoded switch-case kullanilmaz.

**Paket:** `internal/ai/pipeline` — `pipeline.go`, `steps.go`, `runner.go`, `registry.go`, `events.go`, `job_repository.go`, `queue.go`

**Adim sirasi (queue job adlari):**

1. `extract_text` — Text Extraction
2. `chunk_document` — Chunk Generation
3. `generate_embedding` — Embedding
4. `generate_summary` — Document Summary
5. `classify_document` — Document Classification (Ollama `llama3.1:8b`, `documents.ai_document_type`)
6. `extract_metadata` — Metadata Extraction (Ollama `llama3.1:8b`, `document_ai_metadata`)
7. `extract_dates` — Important Date Extraction (AI suggestions on `document_important_dates`)
8. `generate_tags` — Auto Tag Suggestion (Ollama `llama3.1:8b`, `document_suggested_tags`)
9. `extract_entities` — Entity Extraction (Ollama `llama3.1:8b`, `document_entities`)
10. `link_related_documents` — Related Document Suggestions (pgvector + boost signals, `document_related_suggestions`)
11. `completed` — Finalize (belge `processed`, `DocumentProcessed` event)

**Step interface:**

```go
type PipelineStep interface {
    Name() string
    Execute(ctx context.Context, job *PipelineJob) error
    CanRun(ctx context.Context, job *PipelineJob) bool
}
```

**PipelineJob:** `tenantId`, `documentId`, `userId`, `uploadId`, `metadata`, `context`

**Registry:** `Register(step)` ile adimlar eklenir; `RegisterDefaultSteps(executor)` mevcut RAG akisini bozmadan text/chunk/embedding/summary adimlarini baglar.

**Queue:** Her adim ayri `pipeline_step` job'u (`jobs.TypePipelineStep`). Worker `ProcessPipelineStepJob` ile tek adim calistirir.

**Durum tablosu:** `document_ai_jobs` — `id`, `tenant_id`, `document_id`, `step_name`, `status`, `started_at`, `completed_at`, `error`, `retry_count`, `created_at`

**Status degerleri:** `pending`, `running`, `completed`, `failed`, `retrying`, `skipped`

**Retry:** Adim basarisiz olursa retry policy (max 2) uygulanir; tek adim hatasi tum pipeline'i durdurmaz, sonraki adima gecilir.

**Events:** `DocumentEmbeddingCompleted`, `DocumentSummaryCompleted`, `DocumentMetadataCompleted`, `DocumentTagsGenerated`, `DocumentEntitiesExtracted`, `DocumentPipelineCompleted`

**API:** `GET /api/v1/documents/:id/ai-progress` — Document Workspace'te AI Progress paneli

**Kurallar:** Mevcut RAG, semantic search, worker ve queue davranisi korunur; tenant isolation ve provider bagimsizligi (`DocumentExecutor` arayuzu) surdurulur. Pipeline runner yoksa legacy monolithic isleme devam eder.

### Tag Management

Enterprise etiket modulu: lifecycle, validation, alias, merge, analytics ve gelismis arama.

**API:**

- `GET /api/v1/tags` — liste (documentCount, lifecycleStatus, icon, color)
- `GET /api/v1/tags/search?q=` — isim, alias, status, kullanim
- `GET /api/v1/tags/:id` — detay
- `GET /api/v1/tags/:id/stats` — kullanim analitigi
- `GET|POST|DELETE /api/v1/tags/:id/aliases` — alias yonetimi
- `POST /api/v1/tags/:id/merge` — etiket birlestirme
- `POST /api/v1/tags` — olustur (member+)
- `PUT /api/v1/tags/:id` — guncelle (owner/admin)
- `DELETE /api/v1/tags/:id` — soft delete (owner/admin)

### Tag Alias

Model: `tag_aliases` (`tenant_id`, `tag_id`, `alias`). Arama ve AI onerisi kabulunde alias cozumlemesi yapilir.

### Tag Merge

`POST /api/v1/tags/:id/merge` body: `targetTagId`. Kaynak etiketin `document_tags` iliskileri hedefe tasinir; duplicate onlenir; kaynak `archived` olur. Audit: `tag_merged`.

### Tag Governance

Lifecycle: `active`, `archived`, `hidden`, `deleted` (soft delete). Validation: tenant bazli case-insensitive unique name, max uzunluk, bos isim engeli, `is_system` koruma.

Izinler: owner/admin CRUD+merge+alias; member goruntuleme, belgeye atama ve yeni tag olusturma.

Audit: `tag_created`, `tag_updated`, `tag_deleted`, `tag_archived`, `tag_merged`, `tag_alias_created`, `tag_alias_deleted`.

Frontend: `TagsPage` — liste, renk/ikon, kullanim sayisi, status badge, detay modal, alias, merge, arsiv/sil.

### Sharing Architecture

Enterprise paylasim modulu: public link, e-posta daveti, internal user ve internal workspace tipleri. Model `document_share_links` ve `document_shares` tablolarinda genisletildi.

Paylasim tipleri: `public_link`, `email_invite`, `internal_user`, `internal_workspace`.

Durumlar (`status`): `active`, `expired`, `revoked`, `disabled` — `EffectiveShareStatus()` runtime'da son kullanma ve erisim limitini degerlendirir.

Mevcut endpoint'ler korunur; yeni endpoint'ler:

- `POST /api/v1/documents/bulk-share-links` — toplu public link
- `GET /api/v1/documents/share-links/:shareLinkId/analytics`
- `POST /api/v1/documents/:id/share-links/:shareLinkId/revoke`

### Public Share

Public linkler token ile erisilir; dosya presigned URL ile verilmez, backend stream eder.

Son kullanma preset'leri: `never`, `1d`, `7d`, `30d`, `custom`. Suresi dolan link otomatik `expired` olur.

Guvenlik: bcrypt ile sifre hash; opsiyonel `max_view_count`; opsiyonel `allowed_domains` (e-posta domain kisiti, `X-Share-Visitor-Email` veya `visitorEmail` query).

Izinler: `allow_download`, `allow_print`, `allow_copy_text` ayri bayraklar. Opsiyonel filigran altyapisi (`watermark_enabled`, `watermark_template`) — render sonraki asamada.

### Internal Share

Tenant uyesi bazli paylasim `document_shares` tablosunda; `share_type=internal_user`. Mevcut CRUD endpoint'leri ve permission kontrolleri korunur.

`POST/GET/PUT/DELETE /api/v1/documents/:id/internal-shares`, `GET /api/v1/documents/shared-with-me`.

### Share Security

Sifre korumasi: `POST /api/v1/public/share/:token/verify-password` — JWT access token doner.

Domain kisiti: izinli domain listesi JSON olarak saklanir; eslesmeyen erisim `share_access_denied`.

Erisim limiti: `max_view_count` asildiginda link `disabled` olur ve `share_limit_reached` audit kaydi olusur.

### Share Analytics

Her link icin: `view_count`, `download_count`, `last_opened_at`, `last_downloaded_at`, `last_ip_address`, `last_user_agent`, `created_by_user_id`.

Audit aksiyonlari: `share_created`, `share_opened`, `share_downloaded`, `share_password_failed`, `share_expired`, `share_revoked`, `share_access_denied`, `share_limit_reached`.

Frontend: `ShareLinkModal` — Genel, Guvenlik, Izinler, Analitik sekmeleri.

### AI Auto Tag Suggestion

Pipeline `generate_tags` adimi belge metninden 5-10 Turkce etiket onerir (`document_suggested_tags`).

API:

- `GET /api/v1/documents/:id/suggested-tags`
- `POST /api/v1/documents/:id/suggested-tags/:suggestedTagId/accept`
- `POST /api/v1/documents/:id/suggested-tags/:suggestedTagId/reject`
- `POST /api/v1/documents/:id/suggested-tags/accept-all`

Kabul: tenant'ta tag yoksa olusturulur, `document_tags` iliskisi eklenir, oneri `accepted` olur. Audit: `ai_tags_generated`, `ai_tag_accepted`, `ai_tag_rejected`.

### AI Metadata Extraction

Pipeline `extract_metadata` adimi belge metninden yapilandirilmis metadata cikarir (`document_ai_metadata`).

API:

- `GET /api/v1/documents/:id/ai-metadata`
- `POST /api/v1/documents/:id/ai-metadata/generate`

Metadata JSON: `documentType`, `title`, `parties`, `dates`, `amounts`, `identifiers`, `addresses`, `keywords`, `summary`. Audit: `ai_metadata_generated`, `ai_metadata_viewed`.

### AI Important Date Extraction

Pipeline `extract_dates` adimi belgeden onemli tarihleri bulur ve `document_important_dates` tablosuna `status=suggested`, `source=ai` olarak kaydeder.

API:

- `GET /api/v1/documents/:id/important-dates/suggestions`
- `POST /api/v1/documents/:id/important-dates/suggestions/:id/accept`
- `POST /api/v1/documents/:id/important-dates/suggestions/:id/reject`
- `POST /api/v1/documents/:id/important-dates/extract-ai`

Accept: `status=active`. Reject: `status=rejected`. Dashboard yalnizca `active` tarihleri gosterir. Audit: `ai_important_dates_extracted`, `ai_important_date_accepted`, `ai_important_date_rejected`.

### AI Document Classification

Pipeline `classify_document` adimi belge turunu tahmin eder ve `documents` tablosundaki `ai_document_type`, `ai_document_type_confidence`, `ai_classified_at` alanlarini gunceller.

API:

- `GET /api/v1/documents/:id/classification`
- `POST /api/v1/documents/:id/classification/generate`

Desteklenen turler: `invoice`, `contract`, `court_decision`, `vehicle_registration`, `identity_document`, `passport`, `insurance_policy`, `warranty_document`, `bank_receipt`, `academic_article`, `hr_document`, `medical_document`, `tax_document`, `certificate`, `other`.

Smart Collections: `invoices`, `contracts`, `identity_documents`, `insurance_policies`. Audit: `ai_document_classified`, `ai_document_classification_viewed`.

### AI Entity Extraction

Pipeline `extract_entities` adimi belgeden onemli varliklari cikarir ve `document_entities` tablosuna kaydeder.

API:

- `GET /api/v1/documents/:id/entities`
- `POST /api/v1/documents/:id/entities/extract-ai`
- `DELETE /api/v1/documents/:id/entities/:entityId`

Entity turleri: `person`, `company`, `organization`, `email`, `phone`, `address`, `tax_number`, `identity_number`, `passport_number`, `license_plate`, `vehicle_vin`, `invoice_number`, `contract_number`, `court_case_number`, `date`, `amount`, `currency`, `product`, `location`, `custom`.

Yeniden cikarimda eski AI entity kayitlari silinir; ayni `entity_type + value` icin tek kayit tutulur. Audit: `ai_entities_extracted`, `ai_entity_deleted`, `ai_entities_viewed`.

### AI Related Documents

Pipeline `link_related_documents` adimi kaynak belgenin chunk embedding ortalamasini kullanarak tenant icindeki benzer belgeleri bulur ve `document_related_suggestions` tablosuna kaydeder.

API:

- `GET /api/v1/documents/:id/related-suggestions`
- `POST /api/v1/documents/:id/related-suggestions/generate`
- `POST /api/v1/documents/:id/related-suggestions/:id/accept`
- `POST /api/v1/documents/:id/related-suggestions/:id/reject`

Semantic matching pgvector uzerinden yapilir; klasor, etiket, AI belge turu, ortak entity, kaynak ve yil sinyalleri boost alir. Accept: `DocumentRelationship` olusturur ve oneri `accepted` olur. Reject: `rejected`. Audit: `ai_related_documents_generated`, `ai_related_document_accepted`, `ai_related_document_rejected`.

### Hybrid Search (Akıllı Arama)

RAG degildir; yalnizca belge bulma deneyimini iyilestirir. AI cevap uretmez.

Endpoint: `GET /api/v1/documents/hybrid-search?q={query}`

Akis: full-text + prefix → query embedding → pgvector semantic chunk search → `document_ai_metadata` JSONB → `document_entities` → tag / klasor / dosya adi / `ai_document_type` sinyalleri → `documentId` bazinda merge ve karma skor.

Skor sinyalleri: `fullTextScore`, `semanticScore`, `metadataScore`, `entityScore`, `tagScore`, `folderScore`, `filenameScore`, `classificationScore`, `favoriteBoost`, `recentBoost`.

`matchType`: `full_text`, `semantic`, `metadata`, `entity`, `hybrid`.

Tenant / permission: yalnizca aktif tenant ve kullanicinin erisebildigi belgeler (`memberDocumentAccessSQL`).

Cache: 30 sn TTL; anahtar `tenantId + userId + query hash`. `InvalidateGlobalSearch` ile `DocumentUpdated`, `TagAssigned`, `FolderUpdated`, `AI_MetadataCompleted`, `AI_EntitiesExtracted`, `AI_ClassificationCompleted` olaylarinda temizlenir.

Mevcut `GET /api/v1/documents/search` ve `GET /api/v1/documents/semantic-search` endpointleri degistirilmez.

Frontend:

- Global Search (Ctrl+K), Belgeler sayfasi ve Arama Sonuclari sayfasinda mod secimi: **Hizli Arama** (full-text) / **Akilli Arama** (hybrid).
- Sonuc kartlarinda `matchType` badge, snippet, etiketler, `aiDocumentType`, matched entity.
- Sonuca tiklaninca `/documents/:id?q={query}`.

### Workspace AI Insights (Aşama 1)

RAG veya tool calling degildir; mevcut structured data uzerinden insight uretir.

Endpoint: `GET /api/v1/ai/workspace-insights`

Insight turleri: yaklasan onemli tarihler, onay bekleyen belgeler, son onemli yuklemeler, AI siniflandirmasi (sozlesme/fatura/sigorta), eksik metadata, islenemeyen belgeler, iliskili belge onerileri, sik kullanilan etiketler, en aktif klasorler.

Response alanlari: `title`, `description`, `type`, `priority`, `relatedDocuments`, `actionLabel`, `actionUrl`.

Tenant / permission: `MemberDocumentScopeSQL` ve onay sorgularinda kullanici kapsami.

Cache: 60 sn TTL; belge, metadata, siniflandirma, etiket, klasor ve workflow eventlerinde invalidation.

Frontend: Dashboard `AI İçgörüleri` paneli.

### AI Action Suggestions

Kullanıcıya uygulanabilir aksiyon önerileri sunar. AI otomatik işlem yapmaz; kabul sonrası frontend mevcut ekranlara yönlendirir.

Model: `ai_action_suggestions` (`AIActionSuggestion`)

Endpointler:

- `GET /api/v1/ai/action-suggestions`
- `GET /api/v1/documents/:id/ai/action-suggestions`
- `POST /api/v1/documents/:id/ai/action-suggestions/generate`
- `POST /api/v1/ai/action-suggestions/:id/accept`
- `POST /api/v1/ai/action-suggestions/:id/reject`

`suggestion_type`: `add_tag`, `move_folder`, `create_reminder`, `start_approval`, `link_related`, `regenerate_metadata`, `generate_summary`.

`status`: `pending`, `accepted`, `rejected`. Accept yalnızca status günceller.

Üretim: mevcut structured data (sınıflandırma, etiketler, klasörler, metadata, özet, ilişkili belgeler, onay, entity tarihleri).

Frontend: Belge workspace `AI Önerileri` paneli; Dashboard içgörülerinde aksiyon önerileri listesi.

### AI Tool Registry (Aşama 1)

Package: `internal/ai/tools`

Tool interface: `Name`, `Description`, `Schema`, `CanExecute`, `Execute`.

Registry: `Register`, `Get`, `List`.

İlk tool'lar: `add_tag_to_document`, `move_document_to_folder`, `create_important_date`, `create_workflow`, `create_document_relationship`, `generate_summary`, `reprocess_document`, `favorite_document`.

`AIActionSuggestion` accept: `action_payload` içinde `toolName` + `input` varsa permission kontrolü sonrası tool execute edilir; başarılıysa `status=accepted`. Hata durumunda öneri pending kalır.

Audit: `ai_tool_executed`, `ai_tool_execution_failed`, `ai_action_suggestion_accepted`, `ai_action_suggestion_rejected`.

Frontend: Kabul Et öncesi onay dialogu; başarı/hata mesajları.

### Workspace Agent Chat

`POST /api/v1/ai/agent-chat` — workspace seviyesinde agent sohbeti (auth gerekli).

Request: `message`, opsiyonel `sessionId`, `documentId`, `folderId`.

Response: `sessionId`, `answer`, `sources`, `suggestedActions` (her biri `suggestionId`, `toolName`, `actionPayload`, `requiresConfirmation: true`).

Davranış: workspace context + RAG chunkları toplanır; Ollama ile cevap üretilir. Aksiyon isteklerinde tool execute edilmez; `AIActionSuggestion` oluşturulur. Kabul `POST /api/v1/ai/action-suggestions/:id/accept` ile yapılır.

Session scope: `workspace` veya `agent` (belge bağlamında).

Audit: `ai_agent_message_sent`, `ai_agent_action_suggested`.

Frontend: AI Asistan sayfasında RAG / Agent Mode toggle; belge workspace AI panelinde agent desteği.

### RAG Chat

`POST /api/v1/ai/chat` — kullanici belgeleriyle sohbet (auth gerekli).

Request: `question`, opsiyonel `sessionId`, `documentId`, `folderId`, `tagIds`, `collectionKey`.

Akis: soru embed edilir → `document_chunks` uzerinde pgvector arama (tenant + permission filtreli) → en alakali 5-8 chunk context → `llama3.1:8b` ile cevap → mesajlar `ai_chat_sessions` / `ai_chat_messages` tablolarina kaydedilir.

Response: `sessionId`, `answer`, `sources[]`.

Context yoksa sabit mesaj: *Bu konuda belgelerinizde yeterli bilgi bulamadım.*

Sohbet gecmisi API:

- `POST /api/v1/ai/chat-sessions`
- `GET /api/v1/ai/chat-sessions` (`scopeType`, `scopeId` filtreli)
- `GET /api/v1/ai/chat-sessions/:id`
- `DELETE /api/v1/ai/chat-sessions/:id`
- `POST /api/v1/ai/chat-sessions/:id/messages`

### Document Summary

- `GET /api/v1/documents/:id/summary` — belgenin son özetini getirir
- `POST /api/v1/documents/:id/summary/generate` — sync özet üretir (queue'ya taşınabilir `SummaryGenerator` arayüzü)

Özet üretimi: chunk'lar (yoksa `document_texts.full_text`) → `llama3.1:8b` → `summary` + `keyPoints`. Belge başına latest kayıt (`document_summaries`).

### Future AI Integration Rules

AI yan etkileri CRUD icinde degil, Extension veya AILogicListener uzerinden event-driven calisacaktir.

### Open Closed Principle

Yeni ozellik eklerken mevcut CRUD degistirilmez; yeni Listener veya Extension eklenir.

### Thin Controllers

Akis: **Controller → Service → Repository → Events → Listeners**

Controller is mantigi icermez; servis domain event publish eder.

