import {
  AlertCircle,
  CalendarDays,
  ClipboardCheck,
  Clock,
  Cloud,
  FileText,
  FileType,
  HardDrive,
  Mail,
  Share2,
  Sheet,
  Star,
  type LucideIcon,
} from "lucide-react";

const iconMap: Record<string, LucideIcon> = {
  clock: Clock,
  "calendar-days": CalendarDays,
  star: Star,
  "share-2": Share2,
  "clipboard-check": ClipboardCheck,
  "hard-drive": HardDrive,
  "alert-circle": AlertCircle,
  "file-text": FileText,
  "file-type": FileType,
  sheet: Sheet,
  mail: Mail,
  cloud: Cloud,
};

export function smartCollectionIcon(name: string): LucideIcon {
  return iconMap[name] ?? FileText;
}
