import { Loader2Icon, SendIcon, SparklesIcon } from "lucide-react";
import React from "react";
import Markdown, { type Components } from "react-markdown";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

// compact markdown styling tuned for chat bubbles (the tailwind typography
// `prose` plugin isn't wired up, and its margins are too large here)
const markdownComponents: Components = {
  p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
  strong: ({ children }) => (
    <strong className="font-semibold">{children}</strong>
  ),
  ul: ({ children }) => (
    <ul className="mb-2 list-disc space-y-1 pl-4 last:mb-0">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="mb-2 list-decimal space-y-1 pl-4 last:mb-0">{children}</ol>
  ),
  a: ({ children, href }) => (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="underline underline-offset-2"
    >
      {children}
    </a>
  ),
  code: ({ children }) => (
    <code className="rounded bg-background/60 px-1 py-0.5 font-mono text-xs">
      {children}
    </code>
  ),
  pre: ({ children }) => (
    <pre className="my-2 overflow-x-auto rounded bg-background/60 p-2 text-xs last:mb-0">
      {children}
    </pre>
  ),
};

type ChatMessage = {
  role: "user" | "assistant";
  content: string;
  toolsUsed?: string[];
};

type ChatApiResponse = {
  message: string;
  toolsUsed?: string[];
};

const SUGGESTIONS = [
  "What's my balance?",
  "Show my last 5 transactions",
  "How many sats is $20?",
];

export function Chat() {
  const [messages, setMessages] = React.useState<ChatMessage[]>([]);
  const [input, setInput] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(false);
  const scrollRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    scrollRef.current?.scrollTo({
      top: scrollRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, [messages, isLoading]);

  async function send(text: string) {
    const trimmed = text.trim();
    if (!trimmed || isLoading) {
      return;
    }

    const nextMessages: ChatMessage[] = [
      ...messages,
      { role: "user", content: trimmed },
    ];
    setMessages(nextMessages);
    setInput("");
    setIsLoading(true);

    try {
      const response = await request<ChatApiResponse>("/api/ai/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: nextMessages.map(({ role, content }) => ({
            role,
            content,
          })),
        }),
      });
      if (!response) {
        throw new Error("No response from assistant");
      }
      setMessages([
        ...nextMessages,
        {
          role: "assistant",
          content: response.message,
          toolsUsed: response.toolsUsed,
        },
      ]);
    } catch (error) {
      toast.error("Couldn't reach the assistant", {
        description: "" + error,
      });
      // drop the optimistic user message so they can retry
      setMessages(messages);
      setInput(trimmed);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex flex-col h-full">
      <AppHeader
        title="Assistant"
        pageTitle="Assistant"
        description="Ask about your balance, transactions, and more. Each message pays a few sats for AI inference from your wallet."
      />

      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto py-6 flex flex-col gap-4"
      >
        {messages.length === 0 && !isLoading && (
          <div className="flex-1 flex flex-col items-center justify-center text-center gap-6">
            <div className="flex flex-col items-center gap-2">
              <SparklesIcon className="w-8 h-8 text-muted-foreground" />
              <p className="text-muted-foreground max-w-sm">
                Your wallet assistant. Ask a question to get started.
              </p>
            </div>
            <div className="flex flex-wrap items-center justify-center gap-2">
              {SUGGESTIONS.map((suggestion) => (
                <Button
                  key={suggestion}
                  variant="outline"
                  size="sm"
                  onClick={() => send(suggestion)}
                >
                  {suggestion}
                </Button>
              ))}
            </div>
          </div>
        )}

        {messages.map((message, index) => (
          <div
            key={index}
            className={cn(
              "flex",
              message.role === "user" ? "justify-end" : "justify-start"
            )}
          >
            <div
              className={cn(
                "rounded-lg px-4 py-2 max-w-[85%] break-words text-sm leading-relaxed",
                message.role === "user"
                  ? "bg-primary text-primary-foreground whitespace-pre-wrap"
                  : "bg-muted"
              )}
            >
              {message.role === "user" ? (
                message.content
              ) : (
                <Markdown components={markdownComponents}>
                  {message.content}
                </Markdown>
              )}
              {!!message.toolsUsed?.length && (
                <div className="mt-2 flex flex-wrap gap-1">
                  {message.toolsUsed.map((tool, toolIndex) => (
                    <span
                      key={toolIndex}
                      className="text-xs text-muted-foreground bg-background rounded px-1.5 py-0.5"
                    >
                      {tool}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}

        {isLoading && (
          <div className="flex justify-start">
            <div className="rounded-lg px-4 py-2 bg-muted text-muted-foreground flex items-center gap-2">
              <Loader2Icon className="w-4 h-4 animate-spin" />
              Thinking…
            </div>
          </div>
        )}
      </div>

      <form
        className="flex items-center gap-2 border-t border-border pt-4"
        onSubmit={(e) => {
          e.preventDefault();
          send(input);
        }}
      >
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask about your wallet…"
          autoFocus
        />
        <Button type="submit" disabled={isLoading || !input.trim()}>
          {isLoading ? (
            <Loader2Icon className="w-4 h-4 animate-spin" />
          ) : (
            <SendIcon className="w-4 h-4" />
          )}
        </Button>
      </form>
    </div>
  );
}
