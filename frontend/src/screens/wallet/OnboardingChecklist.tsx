import { ChevronRightIcon, CircleCheckIcon, CircleIcon } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useOnboardingData } from "src/hooks/useOnboardingData";
import { cn } from "src/lib/utils";

interface ChecklistItemProps {
  title: string;
  checked: boolean;
  description: string;
  to: string;
  disabled: boolean;
  index: number;
}

function OnboardingChecklist() {
  const { isLoading, checklistItems } = useOnboardingData();

  if (isLoading || !checklistItems.find((x) => !x.checked)) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Get started with your Alby Hub</CardTitle>
        <CardDescription>
          Follow these initial steps to set up and make the most of your Alby
          Hub.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col">
        {checklistItems.map((item, index) => (
          <ChecklistItem
            key={item.title}
            index={index}
            title={item.title}
            description={item.description}
            checked={item.checked}
            to={item.to}
            disabled={item.disabled}
          />
        ))}
      </CardContent>
    </Card>
  );
}

function ChecklistItem({
  title,
  checked = false,
  description,
  to,
  disabled = false,
  index,
}: ChecklistItemProps) {
  const content = (
    <div
      className={cn(
        "flex flex-col p-3 relative group rounded-lg",
        !checked && !disabled && "hover:bg-muted",
        disabled && "opacity-50"
      )}
    >
      {!checked && !disabled && (
        <div className="absolute top-0 left-0 w-full h-full items-center justify-end pr-1.5 hidden group-hover:flex opacity-25">
          <ChevronRightIcon className="size-8" />
        </div>
      )}
      <div className="flex items-center gap-2">
        {checked ? (
          <CircleCheckIcon className="size-5" />
        ) : (
          <CircleIcon className="size-5" />
        )}
        <div
          className={cn(
            "text-sm font-medium leading-none",
            checked && "line-through"
          )}
        >
          {index + 1}. {title}
        </div>
      </div>
      {!checked && (
        <div className="text-muted-foreground text-sm mx-7">{description}</div>
      )}
    </div>
  );

  return checked || disabled ? content : <Link to={to}>{content}</Link>;
}

export default OnboardingChecklist;
