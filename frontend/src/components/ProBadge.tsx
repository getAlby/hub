import { SparklesIcon } from "lucide-react";

export function ProBadge() {
  return (
    <span className="inline-flex items-center justify-center rounded-full border border-primary/30 bg-primary/10 px-1.5 py-0.5 text-[0.625rem] text-primary font-bold shrink-0">
      <SparklesIcon className="h-2.5 w-2.5 mr-0.5" />
      PRO
    </span>
  );
}
