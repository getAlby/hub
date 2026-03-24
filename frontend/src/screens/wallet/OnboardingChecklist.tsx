import { CircleCheckIcon, CircleIcon } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { useOnboardingData } from "src/hooks/useOnboardingData";
import { cn } from "src/lib/utils";

interface ChecklistItemProps {
  title: string;
  checked: boolean;
  to: string;
  disabled: boolean;
  index: number;
}

function OnboardingChecklist() {
  const { isLoading, checklistItems } = useOnboardingData();

  if (isLoading || !checklistItems.find((x) => !x.checked)) {
    return null;
  }

  const completedSteps = checklistItems.filter((item) => item.checked).length;
  const totalSteps = checklistItems.length;
  const progressValue =
    totalSteps === 0 ? 0 : (completedSteps / totalSteps) * 100;

  return (
    <Card className="overflow-hidden rounded-[14px] shadow-none">
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">Get Started</CardTitle>
        <CardDescription>
          Follow these steps to set up and make the most of your Alby Hub
        </CardDescription>
      </CardHeader>
      <CardContent className="px-6 pt-0">
        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-2">
            <Progress value={progressValue} />
            <p className="text-xs text-muted-foreground">
              {completedSteps} of {totalSteps} complete
            </p>
          </div>
          <div className="flex flex-col gap-4">
            {checklistItems.map((item, index) => (
              <ChecklistItem
                key={item.title}
                index={index}
                title={item.title}
                checked={item.checked}
                to={item.to}
                disabled={item.disabled}
              />
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function ChecklistItem({
  title,
  checked = false,
  to,
  disabled = false,
  index,
}: ChecklistItemProps) {
  const content = (
    <div
      className={cn(
        "flex items-center gap-2 text-sm leading-5",
        disabled && "opacity-50"
      )}
    >
      {checked ? (
        <CircleCheckIcon className="size-4 shrink-0 text-primary" />
      ) : (
        <CircleIcon className="size-4 shrink-0 text-foreground" />
      )}
      <div
        className={cn(
          "font-medium text-foreground",
          checked && "text-muted-foreground line-through"
        )}
      >
        {index + 1}. {title}
      </div>
    </div>
  );

  return checked || disabled ? (
    content
  ) : (
    <Link to={to} className="rounded-md transition-opacity hover:opacity-80">
      {content}
    </Link>
  );
}

export default OnboardingChecklist;
