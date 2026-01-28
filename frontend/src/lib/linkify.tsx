import React from "react";

const URL_REGEX =
  /(https?:\/\/[^\s]+|(?:www\.)?[a-zA-Z0-9][-a-zA-Z0-9]*\.[a-zA-Z]{2,}(?:\/[^\s]*)?)/gi;
export function linkifyText(text: string): React.ReactNode {
  if (!text) {
    return text;
  }

  const parts: React.ReactNode[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  URL_REGEX.lastIndex = 0;

  while ((match = URL_REGEX.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }

    const url = match[0];
    const href = url.toLowerCase().startsWith("http") ? url : `https://${url}`;

    parts.push(
      <a
        key={match.index}
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="text-primary underline hover:no-underline"
        onClick={(e) => e.stopPropagation()}
      >
        {url}
      </a>
    );

    lastIndex = match.index + url.length;
  }

  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }

  if (parts.length === 0) {
    return text;
  }

  return <>{parts}</>;
}
