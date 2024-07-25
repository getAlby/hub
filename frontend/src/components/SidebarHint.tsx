import { ListTodo, LucideIcon, Zap } from "lucide-react";
import { ReactElement } from "react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { useOnboardingData } from "src/hooks/useOnboardingData";
import useChannelOrderStore from "src/state/ChannelOrderStore";

function SidebarHint() {
  const { isLoading, checklistItems } = useOnboardingData();
  const { order } = useChannelOrderStore();
  const location = useLocation();

  // User has a channel order
  if (
    !location.pathname.startsWith("/channels/") &&
    order &&
    order.status !== "pay"
  ) {
    return (
      <SidebarHintCard
        icon={Zap}
        title="New Channel"
        description="You're currently opening a new channel"
        buttonText="View Channel"
        buttonLink="/channels/order"
      />
    );
  }

  const openChecklistItems = checklistItems.filter((x) => !x.checked);
  if (
    !location.pathname.startsWith("/home") &&
    !location.pathname.startsWith("/channels/order") &&
    !location.pathname.startsWith("/channels/first") &&
    !isLoading &&
    openChecklistItems.length
  ) {
    return (
      <SidebarHintCard
        icon={ListTodo}
        title="Finish Setup"
        description={
          <>
            <Progress
              className="mt-2"
              value={
                ((checklistItems.length - openChecklistItems.length) /
                  checklistItems.length) *
                100
              }
            />
            <div className="text-xs mt-2">
              {checklistItems.length - openChecklistItems.length} out of{" "}
              {checklistItems.length} completed
            </div>
          </>
        }
        buttonText="See Next Steps"
        buttonLink="/home"
      />
    );
  }
}

type SidebarHintCardProps = {
  title: string;
  description: string | ReactElement;
  buttonText: string;
  buttonLink: string;
  icon: LucideIcon;
};
function SidebarHintCard({
  title,
  description,
  icon: Icon,
  buttonText,
  buttonLink,
}: SidebarHintCardProps) {
  return (
    <div className="md:m-4">
      <Card>
        <CardHeader className="p-4">
          <Icon className="h-8 w-8 mb-4" />
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent className="p-4 pt-0">
          <Link to={buttonLink}>
            <Button size="sm" className="w-full">
              {buttonText}
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}

export default SidebarHint;
