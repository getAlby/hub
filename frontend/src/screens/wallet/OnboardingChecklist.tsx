// src/components/OnboardingChecklist.tsx

import { ChevronRight, Circle, CircleCheck } from "lucide-react";
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
}

function OnboardingChecklist() {
  const { isLoading, checklistItems } = useOnboardingData();

  if (isLoading) {
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
        {checklistItems.map((item) => (
          <ChecklistItem
            key={item.title}
            title={item.title}
            description={item.description}
            checked={item.checked}
            to={item.to}
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
}: ChecklistItemProps) {
  const content = (
    <div
      className={cn(
        "flex flex-col p-3 relative group rounded-lg",
        !checked && "hover:bg-muted"
      )}
    >
      {!checked && (
        <div className="absolute top-0 left-0 w-full h-full items-center justify-end pr-1.5 hidden group-hover:flex opacity-25">
          <ChevronRight className="w-8 h-8" />
        </div>
      )}
      <div className="flex items-center gap-2">
        {checked ? (
          <CircleCheck className="w-5 h-5" />
        ) : (
          <Circle className="w-5 h-5" />
        )}
        <div
          className={cn(
            "text-sm font-medium leading-none",
            checked && "line-through"
          )}
        >
          {title}
        </div>
      </div>
      {!checked && (
        <div className="text-muted-foreground text-sm ml-7">{description}</div>
      )}
    </div>
  );

  return checked ? content : <Link to={to}>{content}</Link>;
}

export default OnboardingChecklist;
