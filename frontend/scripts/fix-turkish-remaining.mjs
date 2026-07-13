import fs from "fs";
import path from "path";

const srcDir = path.resolve("src");

const fixes = [
  ["Kisi paylaşimlari yüklenemedi.", "Kişi paylaşımları yüklenemedi."],
  ["Paylaşilan belgeler yüklenemedi.", "Paylaşılan belgeler yüklenemedi."],
  ["Paylaşilan belgeler yükleniyor...", "Paylaşılan belgeler yükleniyor..."],
  ["Belgeyi görüntüleyebilir, indirebilir ve duzenleyebilir.", "Belgeyi görüntüleyebilir, indirebilir ve düzenleyebilir."],
  ["Belgeyi görüntüleyebilir, indirebilir, duzenleyebilir ve baska kullanicilarla paylaşabilir.", "Belgeyi görüntüleyebilir, indirebilir, düzenleyebilir ve başka kullanıcılarla paylaşabilir."],
  ["Oturum suresi doldu. Lutfen tekrar giris yapin.", "Oturum süresi doldu. Lütfen tekrar giriş yapın."],
  ["Bir kullanici sizinle belge paylaştiginda burada görünecek.", "Bir kullanıcı sizinle belge paylaştığında burada görünecek."],
  ["Dokuman içeriklerinde ara...", "Döküman içeriklerinde ara..."],
  ["En az 2 karakter yazin", "En az 2 karakter yazın"],
  ["Dokuman içeriklerinde arama yapmak için sorgunuzu biraz daha uzatin.", "Döküman içeriklerinde arama yapmak için sorgunuzu biraz daha uzatın."],
  ["Arama yapin", "Arama yapın"],
  ["Dokuman içeriklerinde aramak için en az 2 karakter yazin.", "Döküman içeriklerinde aramak için en az 2 karakter yazın."],
  ["Onaylarim", "Onaylarım"],
  ["Bekleyen onayin yok", "Bekleyen onayın yok"],
  ["Adim ${item.stepOrder}", "Adım ${item.stepOrder}"],
  ["Adim ${item.stepOrder}`", "Adım ${item.stepOrder}`"],
  ["Adim ${myCurrentStep.stepOrder}", "Adım ${myCurrentStep.stepOrder}"],
  ["Adim ${step.stepOrder}", "Adım ${step.stepOrder}"],
  ["Adim ekle", "Adım ekle"],
  [" kalici olarak silinsin mi?", " kalıcı olarak silinsin mi?"],
  [" cop kutusuna tasinsin mi?", " çöp kutusuna taşınsın mı?"],
  ["En az 2 karakter yazin.", "En az 2 karakter yazın."],
  ["Geri Yukle", "Geri Yükle"],
  ["Kalici Sil", "Kalıcı Sil"],
  ["Çöp Kutusuna Tasi", "Çöp Kutusuna Taşı"],
  ["Vazgec", "Vazgeç"],
  ["Aktivite Gecmisi", "Aktivite Geçmişi"],
  ["onemli kullanici işlemlerini izleyin.", "önemli kullanıcı işlemlerini izleyin."],
  ["Simdi Senkronize Et", "Şimdi Senkronize Et"],
  ["Kayit yok", "Kayıt yok"],
  ["Ornek: Sozlesme onay süreci", "Örnek: Sözleşme onay süreci"],
  ["Yukleniyor...", "Yükleniyor..."],
  ["Yukleniyor", "Yükleniyor"],
  ["Dosyayi Yukle", "Dosyayı Yükle"],
  ["\"Yukle\"", "\"Yükle\""],
  [">Yukle<", ">Yükle<"],
  ["Kisi paylaşimlari yüklenemedi.", "Kişi paylaşımları yüklenemedi."],
  ["Lutfen bir kullanici seçin.", "Lütfen bir kullanıcı seçin."],
  ["Belge kullaniciyla paylaşildi.", "Belge kullanıcıyla paylaşıldı."],
  ["Paylaşilabilecek baska tenant uyesi bulunamadı.", "Paylaşılabilecek başka tenant üyesi bulunamadı."],
  ["Belge gecmisi ve geri yükleme", "Belge geçmişi ve geri yükleme"],
  ["Yeni Versiyon Yukle", "Yeni Versiyon Yükle"],
  [" etiketini kaldir", " etiketini kaldır"],
  [" etiketi silinsin mi?", " etiketi silinsin mi?"],
  [" klasöru silinsin mi?", " klasörü silinsin mi?"],
  ["${document.original_file_name} sec`", "${document.original_file_name} seç`"],
  [" sec`", " seç`"],
  ["aria-label=\"Ac\"", "aria-label=\"Aç\""],
  ["label: \"Ac\"", "label: \"Aç\""],
  [">Ac<", ">Aç<"],
  ["Isleniyor...", "İşleniyor..."],
  ["Metni Yeniden Isle", "Metni Yeniden İşle"],
  ["Sizden onay bekleniyor (Adim ", "Sizden onay bekleniyor (Adım "],
  ["Onay Sureci Olustur", "Onay Süreci Oluştur"],
  ["Surec Olustur", "Süreç Oluştur"],
  ["Olusturuluyor...", "Oluşturuluyor..."],
  ["Olustur", "Oluştur"],
  ["Tekrar Dene", "Tekrar Dene"],
  ["paylaşim kaldırılsin mi?", "paylaşım kaldırılsın mı?"],
  ["dogrulandi", "doğrulandı"],
  ["Paylaşim kanalı kullanıldı", "Paylaşım kanalı kullanıldı"],
  ["acildi", "açıldı"],
  ["Entegrasyon hatasi", "Entegrasyon hatası"],
  ["Yorum cevabi", "Yorum cevabı"],
  ["Onay istegi", "Onay isteği"],
  ["Onaylandi", "Onaylandı"],
  ["Onay iptal", "Onay iptal"],
  ["Okunmamis", "Okunmamış"],
  ["okunmamis", "okunmamış"],
  ["isaretlenemedi", "işaretlenemedi"],
  ["sayisi alinamadi", "sayısı alınamadı"],
  ["kaydedilemedi", "kaydedilemedi"],
  ["Belge Detay", "Belge Detayı"],
  ["Tum Belgeler", "Tüm Belgeler"],
  ["Tumunu Gor", "Tümünü Gör"],
  ["Tumunu Okundu Yap", "Tümünü Okundu Yap"],
  ["Dosya sec", "Dosya seç"],
  ["surukle birak", "sürükle bırak"],
  ["basariyla yuklendi", "başarıyla yüklendi"],
  ["basarisiz oldu", "başarısız oldu"],
  ["Etiketler yuklenemedi", "Etiketler yüklenemedi"],
  ["Aktivite kayıtları yuklenemedi", "Aktivite kayıtları yüklenemedi"],
  ["henuz aktivite", "henüz aktivite"],
  ["Cop kutusunda", "Çöp kutusunda"],
  ["Cop Kutusuna", "Çöp Kutusuna"],
  ["cop kutusuna", "çöp kutusuna"],
];

function walk(dir, files = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name !== "node_modules") walk(full, files);
    } else if (/\.(tsx?|ts)$/.test(entry.name)) {
      files.push(full);
    }
  }
  return files;
}

for (const file of walk(srcDir)) {
  let content = fs.readFileSync(file, "utf8");
  const original = content;
  for (const [from, to] of fixes) {
    if (content.includes(from)) {
      content = content.split(from).join(to);
    }
  }
  if (content !== original) {
    fs.writeFileSync(file, content, "utf8");
    console.log("fixed:", path.relative(srcDir, file));
  }
}
