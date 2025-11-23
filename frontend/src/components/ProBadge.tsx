import { Sparkles } from "lucide-react";

export function ProBadge() {
  return (
    <span className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-amber-300 to-amber-500 px-1.5 py-0.5 text-[0.625rem] text-black font-bold shadow-sm shrink-0">
      <Sparkles className="h-2.5 w-2.5 mr-0.5" />
      PRO
    </span>
  );
}
