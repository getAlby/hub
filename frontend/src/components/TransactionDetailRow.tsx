import { CopyIcon } from "lucide-react";
import { Button } from "src/components/ui/button";
import { copyToClipboard } from "src/lib/clipboard";

type TransactionDetailRowProps = {
  label: React.ReactNode;
  children: React.ReactNode;
  // when set, the value is shown next to a copy button that copies this string
  copyable?: string;
};

export function TransactionDetailRow({
  label,
  children,
  copyable,
}: TransactionDetailRowProps) {
  return (
    <div>
      <p>{label}</p>
      <div className="flex items-center justify-between gap-4">
        {/* break-all lives on the wrapper so embedded JSX (links, spans) wraps too */}
        <div className="text-muted-foreground break-all">{children}</div>
        {copyable !== undefined && (
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="text-muted-foreground"
            onClick={() => copyToClipboard(copyable)}
          >
            <CopyIcon />
          </Button>
        )}
      </div>
    </div>
  );
}
