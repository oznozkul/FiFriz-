import { FileText } from "lucide-react";
import { Link } from "react-router-dom";
import { EmptyState, PageTitle, PrimaryButton } from "../components/common";

export function SummariesPage() {
  return (
    <div className="space-y-6">
      <PageTitle title="Özetler" subtitle="AI ile oluşturulan belge özetlerinize hızlı erişim." />
      <EmptyState
        title="Özetler belge detayında"
        description="Her belge için oluşturulan AI özetini belge workspace'inin sağ panelindeki «AI Özeti» bölümünden görüntüleyebilirsiniz."
        action={
          <Link to="/documents">
            <PrimaryButton>
              <FileText className="h-4 w-4" />
              Belgelere git
            </PrimaryButton>
          </Link>
        }
      />
    </div>
  );
}
