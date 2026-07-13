# FiFriz Design System

Bu doküman FiFriz frontend geliştirmelerinde tek kaynak tasarım rehberidir. Tüm yeni ekranlar, bileşenler ve UX kararları bu kurallara uymalıdır.

**Teknoloji:** React + TypeScript + Tailwind CSS  
**Ortak bileşenler:** `frontend/src/components/common/`  
**Tasarım token'ları:** `frontend/tailwind.config.js`, `frontend/src/styles.css`

---

## 1. Product Design Principles

FiFriz bir enterprise doküman yönetim ürünüdür; arayüz güçlü ama ağır olmamalıdır.

| İlke | Açıklama |
|------|----------|
| **Sade** | Her ekranda tek birincil aksiyon. Gereksiz çerçeve, renk ve metin yok. |
| **Profesyonel** | Kurumsal kullanıcıya güven veren düzen; tutarlı hizalama ve tipografi. |
| **Hızlı** | Az tıklama, anında geri bildirim, skeleton/loading ile algılanan performans. |
| **Kullanıcı dostu** | Türkçe, net etiketler; teknik jargon kullanıcıya gösterilmez. |
| **Enterprise ama karmaşık değil** | Gelişmiş özellikler (workflow, paylaşım, AI) katmanlı sunulur; varsayılan görünüm sade kalır. |

**UX kararları:**

- Birincil işlem her zaman görünür ve violet ile vurgulanır.
- İkincil işlemler `secondary` veya `ghost` butonla sunulur.
- Tehlikeli işlemler onay diyaloğu gerektirir (`useConfirm`).
- Başarı/hata her zaman toast veya inline mesajla bildirilir; sessiz başarısızlık yok.

---

## 2. Visual Style

### Genel görünüm

- **Arka plan:** Beyaz ve slate tabanlı (`bg-white`, `bg-slate-50/70`, body: `#f8fafc` + hafif violet radial gradient).
- **Accent:** Violet (`#7367f0`) — birincil aksiyon, aktif sekme, focus halkası.
- **Kenarlık:** Yumuşak, düşük kontrast (`border-line` → `#edf0f5`).
- **Kartlar:** `rounded-2xl` (yeni ekranlar); mevcut `.panel` sınıfı `rounded-lg` kullanır — yeni geliştirmelerde `rounded-2xl` tercih edilir.
- **Gölge:** Hafif, yayvan (`shadow-panel`, `shadow-lift`); kartlar havada “yüzer”, agresif drop-shadow kullanılmaz.
- **Spacing:** Ferah; içerik sıkıştırılmaz. Bölümler arası `space-y-4` / `space-y-5` varsayılandır.

### Yüzey hiyerarşisi

1. **Sayfa zemini** — `bg-slate-50/70` veya `bg-white` (ayarlar)
2. **Kart / panel** — `bg-white border border-line shadow-panel rounded-2xl`
3. **Vurgulu alan** — `bg-violet-soft/50 border-violet-ring` (oluşturulan link, seçili öğe)
4. **İkincil zemin** — `bg-slate-50/70` (form bölümleri, filtre alanları)

### İkonlar

- **Lucide React** kullanılır.
- Standart boyut: `h-4 w-4` (buton içi), `h-5 w-5` (toast), `h-6 w-6` (empty state).
- İkon tek başına anlam taşımaz; yanında metin veya `aria-label` olmalı.

---

## 3. Typography

**Font ailesi:** Inter (`font-sans`)

| Rol | Sınıf | Kullanım |
|-----|-------|----------|
| Sayfa başlığı | `text-2xl md:text-3xl font-bold text-ink` | `PageTitle` |
| Bölüm başlığı | `text-lg font-bold text-ink` | Modal başlığı, panel başlığı |
| Alt başlık / kart başlığı | `text-sm font-bold text-ink` | Form bölümü, liste başlığı |
| Gövde metni | `text-sm text-slate-600` veya `text-slate-700` | Açıklamalar, paragraflar |
| İkincil / yardımcı | `text-xs text-slate-400` veya `text-slate-500` | Meta bilgi, tarih, sayaç |
| Tablo başlığı | `text-xs font-semibold uppercase tracking-wide text-slate-400` | Sütun başlıkları |
| Tablo hücresi | `text-sm font-semibold text-ink` (birincil), `text-sm text-slate-500` (ikincil) | Satır içeriği |
| Form label | `text-sm font-semibold text-slate-600` | Her input üstünde |
| Form yardım metni | `text-xs text-slate-500` | Label altı açıklama |
| Badge / etiket | `text-xs font-semibold` | `Badge`, `StatusBadge` |
| Kbd / kısayol | `text-[11px] font-semibold text-slate-400` | Command palette, arama |

**Kurallar:**

- Başlık hiyerarşisi atlanmaz (h1 → h2 → h3).
- Uzun metin `max-w-prose` veya `max-w-md` ile sınırlanır (empty state, modal).
- `font-bold` yalnızca başlık ve vurgulu etiketlerde; gövde metni `font-semibold` veya normal.

---

## 4. Color System

Tüm renkler Tailwind extend token'ları ve semantik anlamlarla kullanılır.

| Token | Değer / sınıf | Kullanım |
|-------|---------------|----------|
| **Primary** | `violet-brand` `#7367f0` | Birincil buton, aktif sekme, link vurgusu |
| **Primary hover** | `violet-hover` `#6758e8` | Primary buton hover |
| **Primary soft** | `violet-soft` `#f1efff` | Hover zemin, seçili satır, ikon kutusu |
| **Primary ring** | `violet-ring` `#e4e0ff` | Focus ring, hover border |
| **Secondary** | `border-line bg-white text-slate-600` | İkincil buton, nötr kart |
| **Success** | `emerald-100` / `emerald-600` | Başarı badge, success toast |
| **Warning** | `amber-100` / `amber-600` | Bekleyen, işleniyor badge |
| **Danger** | `red-600`, `rose-100` / `rose-600` | Silme, hata, red badge |
| **Info** | `sky-50` / `sky-800` | Bilgi toast |
| **Muted** | `muted` `#64748b`, `text-slate-500` | İkincil metin |
| **Ink** | `ink` `#111827` | Birincil metin, başlıklar |
| **Border** | `line` `#edf0f5` | Kart, input, ayırıcı çizgiler |
| **Background** | `bg-white`, `bg-slate-50`, body `#f8fafc` | Sayfa ve panel zeminleri |

**Kurallar:**

- Violet dışı accent renk yalnızca semantik durumlar için (success/warning/danger).
- Ham hex veya rastgele Tailwind renkleri eklenmez; önce bu tabloya uygun token seçilir.
- Metin kontrastı WCAG AA hedeflenir (`text-ink` / `text-slate-600` beyaz zemin üzerinde).

---

## 5. Spacing & Layout

### Sayfa düzeni

```
┌─────────────┬──────────────────────────────────────┐
│  Sidebar    │  Topbar (h-16)                       │
│  w-64       ├──────────────────────────────────────┤
│  (collapsed │  Main content                        │
│   w-[4.5rem])│  padding: responsive                │
└─────────────┴──────────────────────────────────────┘
```

| Öğe | Değer |
|-----|-------|
| **Sidebar genişlik** | `w-64` (256px) açık; `w-[4.5rem]` (72px) daraltılmış |
| **Topbar yükseklik** | `h-16` (64px), `sticky top-0 z-30` |
| **Main padding** | `px-4 pt-4 pb-5` → `sm:px-5` → `md:px-7 md:pt-5 md:pb-8` |
| **Ayarlar sayfası padding** | `px-4 sm:px-6 md:px-8` |
| **Kart iç padding** | `p-4` (kompakt), `p-5` / `p-6` (standart panel) |
| **Grid gap** | `gap-4` (form), `gap-5` / `gap-6` (sayfa bölümleri) |
| **Bölüm aralığı** | `space-y-4` (form), `space-y-5` (sayfa) |
| **Belge workspace sağ panel** | `w-[360px]` (desktop), drawer (mobile) |

### Breakpoint'ler (Tailwind varsayılan)

- `sm` 640px — breadcrumb görünür, padding artar
- `md` 768px — command palette trigger, 2 kolon grid
- `lg` 1024px — sidebar sabit, mobil menü gizlenir

### Container

- Liste/tablo sayfaları tam genişlik (`w-full min-w-0`).
- Modal: `max-w-sm` / `max-w-md` / `max-w-lg` / `max-w-2xl` / `max-w-3xl` içerik yoğunluğuna göre.
- Topbar arama: `max-w-xl` ortalanmış.

---

## 6. Components

Yeni geliştirmelerde önce `frontend/src/components/common/` altındaki bileşenler kullanılır; tekrarlayan stil yazılmaz.

### Button

**Dosya:** `Button.tsx`, `PrimaryButton.tsx`, `IconButton.tsx`

| Varyant | Kullanım |
|---------|----------|
| `primary` | Birincil aksiyon (kaydet, oluştur, yükle) |
| `secondary` | İkincil aksiyon (iptal, filtre, tekrar dene) |
| `danger` | Silme, iptal onayı |
| `ghost` | Toolbar, düşük öncelikli aksiyon |

**Boyut:** `h-10`, `px-4`, `text-sm font-semibold`, `rounded-lg`  
**Durumlar:** `disabled:opacity-60`, `active:scale-[0.98]`, `focus:ring-4 focus:ring-violet-ring/70`  
**Yükleme:** İçerik `Spinner size="sm"` ile değiştirilir; buton `disabled` kalır.

### Input

**Standart metin input:**

```
h-10 w-full rounded-lg border border-line bg-white px-3 text-sm
outline-none focus:border-violet-brand
```

- Hata durumu: `border-red-300 focus:border-red-500`
- Disabled: `bg-slate-50 text-slate-400 cursor-not-allowed`

### SearchInput

**Dosya:** `SearchInput.tsx`

- Sol `Search` ikonu, opsiyonel kbd kısayolu (`Ctrl+K`).
- Wrapper: `h-10 rounded-lg border border-line focus-within:border-violet-brand focus-within:ring-4 focus-within:ring-violet-ring/70`.

### Select

Native `<select>` ile aynı input yüksekliği ve border kuralları:

```
h-10 rounded-lg border border-line bg-white px-3 font-semibold text-slate-700
focus:border-violet-brand
```

### Textarea

```
min-h-24 rounded-lg border border-line bg-white px-3 py-2 text-sm
focus:border-violet-brand
```

### Modal

**Dosya:** `AppModal.tsx` (standart), özel modallar aynı yapıyı izler.

- Overlay: `fixed inset-0 z-[90] bg-slate-900/40 p-4`
- Kutu: `rounded-xl border border-line bg-white shadow-panel`
- Başlık: `border-b px-5 py-4`, `text-lg font-bold text-ink`
- İçerik: `px-5 py-4`
- Footer: `border-t px-5 py-4`, sağa hizalı butonlar
- `role="dialog"`, `aria-modal="true"`, `aria-labelledby`
- Overlay tıklama ve ESC ile kapanır

### Drawer

Tam ekran alt/yan panel pattern (belge workspace mobile):

- Overlay: `bg-black/40` veya `bg-slate-900/40`
- Panel: `bg-white`, sağdan kayar, `w-full` veya sabit genişlik
- Kapat: overlay tıklama, `aria-label="Paneli kapat"`, üst kapat butonu

### Table

**Dosya:** `DocumentTable.tsx`, `DocumentListPanel.tsx`, `TablePagination.tsx`

- Grid tabanlı satır: `grid` + sabit kolon şablonu
- Başlık satırı: üst border, `text-xs font-semibold uppercase text-slate-400`
- Satır: `px-5 py-4`, hover `hover:bg-slate-50/80`
- Sayfalama: tablo altında `TablePagination`
- Yükleme: `TableRowSkeleton`

### Card

- Standart: `.panel` veya `rounded-2xl border border-line bg-white shadow-panel p-5`
- İstatistik kartı: üst ikon + label + değer (`StatCardSkeleton` referans)
- Tıklanabilir kart: `soft-hover` veya `hover:border-violet-ring transition`

### Badge

**Dosya:** `Badge.tsx`

| Tone | Kullanım |
|------|----------|
| `slate` | Nötr, arşiv, pasif |
| `violet` | Kilit, özel durum |
| `green` | Başarı, hazır, onaylı |
| `amber` | Bekliyor, işleniyor |
| `rose` | Hata, red |
| `blue` | Bilgi, entegrasyon |

Boyut: `text-xs font-semibold rounded-md px-2 py-1`

### Toast

**Dosya:** `ToastProvider.tsx`, `services/toast.ts`

- Konum: sağ alt `bottom-4 right-4`, `z-[200]`
- Tipler: `success`, `error`, `info`, `loading`
- Animasyon: `.toast-enter` (0.2s ease-out)
- `aria-live="polite"`; hata için `role="alert"`

### Tooltip

Henüz ortak bileşen yok. Gerektiğinde:

- `text-xs bg-ink text-white rounded-md px-2 py-1`
- Sadece ikon butonlarda zorunlu değilse `aria-label` tercih edilir

### Dropdown

**Dosya:** `SearchDropdown.tsx`, `UserMenu.tsx`, `ContextMenu.tsx`

- Menü: `rounded-lg border border-line bg-white shadow-panel py-1`
- Öğe: `px-3 py-2 text-sm hover:bg-violet-soft/40`
- Aktif: `text-violet-brand font-semibold`

### EmptyState

**Dosya:** `EmptyState.tsx`

- `panel` içinde ortalanmış
- İkon kutusu: `h-12 w-12 rounded-lg bg-violet-soft text-violet-brand`
- Başlık: `text-base font-bold text-ink`
- Açıklama: `text-sm text-slate-500 max-w-md`
- Aksiyon: `mt-5`, `PrimaryButton`

### LoadingSkeleton

**Dosya:** `Skeleton.tsx`

- Temel: `animate-pulse rounded bg-slate-100`
- Hazır şablonlar: `StatCardSkeleton`, `TableRowSkeleton`, `ChatMessageSkeleton`

### ConfirmationDialog

**Dosya:** `ConfirmProvider.tsx` + `useConfirm()`

```ts
const ok = await confirm({
  title: "Paylaşımı kaldır",
  message: "Bu işlem geri alınamaz.",
  confirmLabel: "Kaldır",
  variant: "danger", // veya "primary"
});
```

- `AppModal` tabanlı; varsayılan iptal: "Vazgeç"
- Tehlikeli işlemlerde `variant: "danger"` zorunlu

---

## 7. Status Badge Rules

**Dosya:** `StatusBadge.tsx` — `resolveStatusBadge()` tek kaynak.

| Durum | Label (TR) | Tone | Not |
|-------|------------|------|-----|
| `processing` | İşleniyor | `amber` | AI / worker işlemi devam ediyor |
| `processed` / `ready` | Hazır | `green` | İşlem tamamlandı |
| `failed` | Başarısız | `rose` | Hata; detay tooltip veya panelde |
| `pending` | Bekliyor | `amber` | Onay / kuyruk |
| `approved` | Onaylandı | `green` | Workflow onayı |
| `rejected` | Reddedildi | `rose` | Workflow reddi |
| `archived` | Arşivlendi | `slate` | Pasif arşiv |
| `locked` | Kilitli | `violet` | Belge/klasör kilidi |
| `expired` | Süresi doldu | `slate` | Paylaşım / tarih |

**Kurallar:**

- Yeni durum eklendiğinde yalnızca `STATUS_CONFIG` güncellenir.
- Bilinmeyen durum: ham string gösterilir, tone `slate`.
- Aynı anlama gelen İngilizce/Türkçe key'ler tek label'a map edilir (`ready` = `processed`).

---

## 8. Empty State Rules

Her boş liste, tablo veya panel şu yapıyı izler:

1. **Kısa başlık** — neyin boş olduğu (`"Henüz belge yok"`)
2. **Açıklama** — neden boş veya kullanıcı ne yapabilir (1–2 cümle)
3. **Birincil aksiyon** — `PrimaryButton` ile net sonraki adım
4. **İkon** (gerekirse) — `EmptyState` varsayılan `FileSearch`; bağlama uygun Lucide ikonu

**Ne zaman EmptyState?**

- Filtre sonucu 0 kayıt
- İlk kullanım (henüz veri yok)
- Yetki nedeniyle içerik yok (açıklayıcı mesaj)

**Ne zaman değil?**

- Sayfa ilk yüklenirken → `LoadingState`
- API hatası → Error kuralları (Bölüm 10)

`PageState` bileşeni loading / error / empty / content geçişini yönetir.

---

## 9. Loading Rules

| Senaryo | Bileşen | Varyant |
|---------|---------|---------|
| Sayfa ilk yükleme | `LoadingState` | `panel` veya `fullscreen` |
| Tablo/liste yükleme | `TableRowSkeleton` × N | Satır sayısı kadar |
| Kart grid yükleme | `StatCardSkeleton` | Dashboard kartları |
| Buton aksiyonu | `Spinner size="sm"` | Buton içinde, `disabled` |
| Inline küçük alan | `Spinner` / `LoadingState` | `inline` |
| Chat / mesaj | `ChatMessageSkeleton` | AI paneli |

**Kurallar:**

- Tam sayfa spinner yalnızca ilk mount'ta; sayfalama/filtrede tablo skeleton tercih edilir.
- `toast.loading` uzun API işlemlerinde; tamamlanınca `toast.success` veya `toast.error`.
- Skeleton gerçek layout'u yansıtmalı (yanlış boyut zıplaması yapılmaz).

---

## 10. Error Rules

| Kural | Uygulama |
|-------|----------|
| Kullanıcı dostu mesaj | `"Paylaşım linkleri yüklenemedi."` — API `error.message` veya fallback |
| Teknik hata gösterme | Stack trace, HTTP kodu, SQL hatası UI'da yok |
| Retry | Liste/sayfa hatalarında `Button variant="secondary"` → "Tekrar dene" |
| Form hatası | Alan altında kısa kırmızı metin veya toast |
| Kritik hata | `EmptyState` + retry (`PageState`) |
| Anlık hata | `toast.error()` |

**API hata mesajı önceliği:**

1. Backend `error.message` (Türkçe, anlamlı)
2. Servis katmanı fallback mesajı
3. Genel: `"Bir hata oluştu. Lütfen tekrar deneyin."`

---

## 11. Navigation Rules

### Sidebar

**Dosya:** `layout/Sidebar.tsx`, `data/sidebarNav.ts`

- Gruplar: ana navigasyon, workspace, ayarlar (mantıksal bölümleme)
- Aktif öğe: `text-violet-brand`, `bg-violet-soft`, sol border veya ikon vurgusu
- Daraltma: localStorage `fifriz_sidebar_collapsed`; `w-[4.5rem]` sadece ikon
- Mobile: overlay drawer, `z-50`

### Topbar

**Dosya:** `layout/Topbar.tsx`

- Sol: mobil menü + `Breadcrumb` (`sm+` görünür)
- Orta: command palette trigger (`md+`), `Ctrl+K`
- Sağ: bildirim zili, ayırıcı, `UserMenu`

### Breadcrumb

- Maksimum 3–4 segment; uzun isimler `truncate`
- Son segment aktif (`text-ink font-semibold`)
- Belge workspace'te dinamik güncellenir (`BreadcrumbContext`)

### Command Palette

**Dosya:** `CommandPalette.tsx`

- `Ctrl+K` / `Cmd+K`
- Arama + hızlı navigasyon + yükleme kısayolu
- Mobile: topbar arama ikonu ile açılır

### User Menu

- Avatar / isim kısaltması
- Profil, ayarlar, çıkış
- `Dropdown` kurallarına uyar

---

## 12. Document Workspace Rules

**Dosya:** `components/documentWorkspace/`

### Layout (desktop)

```
┌────────────────────────────────────────┬──────────────┐
│  Toolbar                               │              │
├────────────────────────────────────────┤  Sağ panel   │
│                                        │  w-[360px]   │
│  Preview alanı (FilePreview)           │  accordion   │
│                                        │  bölümler    │
└────────────────────────────────────────┴──────────────┘
```

### Preview alanı

- PDF/görsel: `iframe` veya `img`, max yükseklik `70vh`
- Desteklenmeyen tür: bilgi metni + indirme (izin varsa)
- Arama vurgusu: sarı/turuncu highlight (docx/xlsx preview)

### Toolbar

**Dosya:** `DocumentWorkspaceToolbar.tsx`

- İndir, paylaş, favori, yeniden işle, panel aç/kapa
- `IconButton` + metin (desktop); yalnız ikon (mobile)

### Sağ bilgi paneli

**Dosya:** `DocumentWorkspacePanel.tsx`

- Accordion bölümler: Genel, Sağlık, Etiketler, Klasör, Paylaşımlar, Workflow, Yorumlar, Versiyonlar, Tarihler, Aktivite, AI…
- URL `?panel=` ile derin link
- Varsayılan açık: `general`

### Metadata / AI panel

- AI içerik yüklenirken skeleton veya `Spinner`
- Boş AI sonucu: EmptyState ("Henüz analiz yok" + generate butonu)
- Uzun JSON/metadata: okunabilir tipografi, `text-sm`, bölümlere ayrılmış

### Comments & Versions

- Kronolojik liste; en yeni üstte veya altta (mevcut pattern korunur)
- Her öğe: avatar/isim, zaman, içerik
- Versiyon: indir / geri yükle ayrı aksiyonlar, tehlikeli işlem onaylı

### Mobile

- Preview tam genişlik
- Sağ panel → drawer (`isDrawer`, overlay)
- Floating "Panel" butonu

---

## 13. Responsive Rules

| Breakpoint | Davranış |
|------------|----------|
| **Desktop** `lg+` | Sidebar sabit, tablo tam grid, sağ panel görünür |
| **Tablet** `md` | 2 kolon grid, command palette, bazı kolonlar gizlenir |
| **Mobile** `< lg` | Sidebar drawer, tablo → kart list veya yatay scroll, panel drawer |

**Drawer kullanımı:**

- Sidebar (mobile)
- Belge workspace sağ panel (mobile)
- Uzun filtre formları (gerekirse)

**Tablo → kart list:**

- `sm` altında kritik kolonlar kalır; diğerleri kart içi meta
- `DocumentRow` kompakt mod veya stacked layout

**Touch:**

- Minimum dokunma hedefi 44×44px (`h-10`, `h-11` butonlar)
- Hover-only bilgi mobile'da görünür alternatifle sunulur

---

## 14. Accessibility Rules

| Kural | Uygulama |
|-------|----------|
| Klavye | Tüm interaktif öğeler `Tab` ile erişilebilir; mantıksal focus sırası |
| Focus | `focus:outline-none focus:ring-4 focus:ring-violet-ring/70` |
| Modal ESC | Overlay ve ESC ile kapanır; focus trap modal içinde |
| `aria-label` | Yalnız ikonlu butonlarda zorunlu (`IconButton`) |
| `aria-modal` | Modallarda `role="dialog" aria-modal="true"` |
| Canlı bölgeler | Toast: `aria-live="polite"`; hata: `role="alert"` |
| Kontrast | `text-ink` / `text-slate-600` beyaz zemin; badge'lerde tone tablosu |
| Form | Her input `label` ile ilişkili (`htmlFor` / sarmalayıcı label) |
| Durum | Yükleme butonunda `disabled` + görsel spinner |

---

## 15. Micro Interaction Rules

| Etkileşim | Değer |
|-----------|-------|
| **Hover** | Border → `violet-ring`; zemin → `violet-soft/40`; metin → `violet-brand` |
| **Active** | Buton `active:scale-[0.98]` veya `active:scale-95` (icon) |
| **Transition** | `transition` (150–200ms); layout shift yaratmayan özellikler |
| **Toast** | `toast-enter` 0.2s ease-out, alttan giriş |
| **Modal** | Overlay fade; içerik scale/fade (mevcut: anında; iyileştirmede 150ms) |
| **Sidebar collapse** | `transition-all duration-200` |
| **Sekme** | Alt border `border-b-2 border-violet-brand` aktif sekme |

**Yapılmaması gerekenler:**

- Uzun (>300ms) animasyonlar
- Her hover'da gölge sıçraması
- Otomatik oynayan dikkat dağıtıcı animasyonlar

---

## Uygulama Kontrol Listesi

Yeni bir ekran veya özellik geliştirirken:

- [ ] Ortak bileşenler kullanıldı mı?
- [ ] Renkler token tablosundan mı?
- [ ] Loading, empty ve error durumları tanımlı mı?
- [ ] Birincil aksiyon tek ve net mi?
- [ ] Tehlikeli işlem `useConfirm` ile korunuyor mu?
- [ ] Mobile layout düşünüldü mü?
- [ ] `aria-label` ve focus ring var mı?
- [ ] Toast ile kullanıcıya geri bildirim veriliyor mu?

---

*Son güncelleme: bu doküman FiFriz mevcut frontend implementasyonu (`tailwind.config.js`, `styles.css`, `components/common`) ile uyumludur. Token veya bileşen değişikliklerinde bu dosya güncellenmelidir.*
