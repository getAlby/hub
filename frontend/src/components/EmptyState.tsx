import { LucideIcon } from "lucide-react";
import React from "react";
import { LinkButton } from "src/components/ui/custom/link-button";
import { cn } from "src/lib/utils";

interface Props {
  icon: LucideIcon;
  title: string;
  description: string;
  buttonText: string;
  buttonLink: string;
  showButton?: boolean;
  showBorder?: boolean;
}

const EmptyState: React.FC<Props> = ({
  icon: Icon,
  title: message,
  description: subMessage,
  buttonText,
  buttonLink,
  showButton = true,
  showBorder = true,
}) => {
  return (
    <div
      className={cn(
        "flex flex-1 items-center justify-center rounded-lg p-8",
        showBorder && "shadow-xs border border-dashed"
      )}
    >
      <div className="flex flex-col items-center gap-1 text-center max-w-sm">
        <Icon className="w-10 h-10 text-muted-foreground" />
        <h3 className="mt-4 text-lg font-semibold">{message}</h3>
        <p className="text-sm text-muted-foreground">{subMessage}</p>
        {showButton && (
          <LinkButton to={buttonLink} className="mt-4">
            {buttonText}
          </LinkButton>
        )}
      </div>
    </div>
  );
};

export default EmptyState;
