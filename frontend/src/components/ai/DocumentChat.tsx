import { MoreHorizontal } from "lucide-react";
import { IconButton } from "../common";
import { ChatBubble } from "./ChatBubble";
import { MessageInput } from "./MessageInput";

export function DocumentChat() {
  return (
    <aside className="panel flex min-h-[640px] flex-col overflow-hidden">
      <div className="flex h-12 items-center justify-between border-b border-line px-4">
        <h2 className="font-bold text-ink">AI Asistan</h2>
        <IconButton className="h-8 w-8 border-0" aria-label="Kapat">
          <MoreHorizontal className="h-4 w-4" />
        </IconButton>
      </div>
      <div className="flex-1 space-y-4 overflow-auto p-4">
        <ChatBubble side="right" text="Bu sözleşmenin bitiş tarihi ne?" />
        <ChatBubble side="left" text="Bu sözleşmenin bitiş tarihi 31.12.2025 olarak görünüyor." source="Kaynak: Sayfa 1" />
        <ChatBubble side="right" text="Depozito tutarı ne kadar?" />
        <ChatBubble side="left" text="Depozito tutarı 15.000 TL olarak belirtilmiş." source="Kaynak: Sayfa 1" />
      </div>
      <div className="p-4">
        <MessageInput compact />
      </div>
    </aside>
  );
}
