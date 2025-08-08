import { LucideIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";

interface Props {
  icon: LucideIcon;
  title: string;
  description: string;
  buttonText: string;
  buttonLink: string;
  showButton?: boolean;
}

const EmptyState: React.FC<Props> = ({
  icon: Icon,
  title: message,
  description: subMessage,
  buttonText,
  buttonLink,
  showButton = true,
}) => {
  return (
    <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-xs p-8">
      <div className="flex flex-col items-center gap-1 text-center max-w-sm">
        <Icon className="w-10 h-10 text-muted-foreground" />
        <h3 className="mt-4 text-lg font-semibold">{message}</h3>
        <p className="text-sm text-muted-foreground">{subMessage}</p>
        {showButton && (
          <Link to={buttonLink}>
            <Button className="mt-4">{buttonText}</Button>
          </Link>
        )}
      </div>
    </div>
  );
};

export default EmptyState;
