import { Loader2 } from "lucide-react";
import { cn } from "src/lib/utils";

function Loading({ className }: { className?: string }) {
  return (
    <>
      <Loader2 className={cn("h-6 w-6 animate-spin", className)}>
        <span className="sr-only">Loading...</span>
      </Loader2>
    </>
  );
}

export default Loading;
