import { AiMessageBubble } from "./AiMessageBubble";

type ChatBubbleProps = {
  side: "left" | "right";
  text: string;
  source?: string;
};

export function ChatBubble({ side, text, source }: ChatBubbleProps) {
  return <AiMessageBubble role={side === "right" ? "user" : "assistant"} text={text} source={source} maxWidthClassName="max-w-[78%]" />;
}
