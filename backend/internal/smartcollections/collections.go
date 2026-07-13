package smartcollections

type CollectionDefinition struct {
	Key         string
	Title       string
	Description string
	Icon        string
}

var definitions = []CollectionDefinition{
	{Key: "recent_7_days", Title: "Son 7 Gün", Description: "Son bir haftada eklenen belgeler.", Icon: "clock"},
	{Key: "uploaded_this_month", Title: "Bu Ay Yüklenenler", Description: "İçinde bulunduğumuz ay yüklenen belgeler.", Icon: "calendar-days"},
	{Key: "favorites", Title: "Favoriler", Description: "Favori olarak işaretlenmiş belgeler.", Icon: "star"},
	{Key: "shared_with_me", Title: "Benimle Paylaşılanlar", Description: "Diğer kullanıcıların sizinle paylaştığı belgeler.", Icon: "share-2"},
	{Key: "pending_approval", Title: "Onay Bekleyenler", Description: "Onay sürecinde bekleyen belgeler.", Icon: "clipboard-check"},
	{Key: "large_files", Title: "Büyük Dosyalar", Description: "50 MB ve üzeri dosyalar veya en büyük 20 belge.", Icon: "hard-drive"},
	{Key: "failed_processing", Title: "İşleme Başarısız", Description: "İşlenirken hata alan belgeler.", Icon: "alert-circle"},
	{Key: "pdf_documents", Title: "PDF Belgeler", Description: "PDF uzantılı belgeler.", Icon: "file-text"},
	{Key: "word_documents", Title: "Word Belgeler", Description: "DOC ve DOCX uzantılı belgeler.", Icon: "file-type"},
	{Key: "excel_documents", Title: "Excel Belgeler", Description: "XLS ve XLSX uzantılı belgeler.", Icon: "sheet"},
	{Key: "email_imported", Title: "E-postadan Gelenler", Description: "IMAP, Gmail veya Outlook ile içe aktarılan belgeler.", Icon: "mail"},
	{Key: "cloud_imported", Title: "Buluttan Gelenler", Description: "Google Drive, OneDrive veya Dropbox kaynaklı belgeler.", Icon: "cloud"},
	{Key: "invoices", Title: "Faturalar", Description: "AI tarafından fatura olarak sınıflandırılan belgeler.", Icon: "receipt"},
	{Key: "contracts", Title: "Sözleşmeler", Description: "AI tarafından sözleşme olarak sınıflandırılan belgeler.", Icon: "file-signature"},
	{Key: "identity_documents", Title: "Kimlik Belgeleri", Description: "AI tarafından kimlik belgesi olarak sınıflandırılan belgeler.", Icon: "id-card"},
	{Key: "insurance_policies", Title: "Sigorta Belgeleri", Description: "AI tarafından sigorta poliçesi olarak sınıflandırılan belgeler.", Icon: "shield"},
}

func AllDefinitions() []CollectionDefinition {
	out := make([]CollectionDefinition, len(definitions))
	copy(out, definitions)
	return out
}

func GetDefinition(key string) (CollectionDefinition, bool) {
	for _, item := range definitions {
		if item.Key == key {
			return item, true
		}
	}
	return CollectionDefinition{}, false
}

func IsValidKey(key string) bool {
	_, ok := GetDefinition(key)
	return ok
}
