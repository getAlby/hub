import { LucideIcon } from "lucide-react";
import React from "react";
import { LinkButton } from "src/components/ui/custom/link-button";
import { cn } from "src/lib/utils";

type Variant = "dashed" | "muted" | "none";

type Props = {
  icon: LucideIcon;
  title: string;
  description: string;
  variant?: Variant;
} & (
  | { buttonText: string; buttonLink: string }
  | { buttonText?: never; buttonLink?: never }
);

const variantClasses: Record<Variant, string> = {
  dashed: "shadow-xs border border-dashed",
  muted: "bg-muted",
  none: "",
};

const EmptyState: React.FC<Props> = ({
  icon: Icon,
  title: message,
  description: subMessage,
  variant = "dashed",
  buttonText,
  buttonLink,
}) => {
  return (
    <div
      className={cn(
        "flex flex-1 items-center justify-center rounded-lg p-8",
        variantClasses[variant]
      )}
    >
      <div className="flex flex-col items-center gap-1 text-center max-w-sm">
        <Icon className="w-10 h-10 text-muted-foreground" />
        <h3 className="mt-4 text-lg font-semibold">{message}</h3>
        <p className="text-sm text-muted-foreground">{subMessage}</p>
        {buttonText && buttonLink && (
          <LinkButton to={buttonLink} className="mt-4">
            {buttonText}
          </LinkButton>
        )}
      </div>
    </div>
  );
};

export default EmptyState;
