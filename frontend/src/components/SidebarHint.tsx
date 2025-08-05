import {
  HeartIcon,
  ListTodoIcon,
  LucideIcon,
  XIcon,
  ZapIcon,
} from "lucide-react";
import React, { ReactElement } from "react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Progress } from "src/components/ui/progress";
import { useToast } from "src/components/ui/use-toast";
import { localStorageKeys, SUPPORT_ALBY_CONNECTION_NAME } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useOnboardingData } from "src/hooks/useOnboardingData";
import useChannelOrderStore from "src/state/ChannelOrderStore";

function SidebarHint() {
  const { isLoading, checklistItems } = useOnboardingData();
  const { data: supportAlbyAppsData } = useApps(undefined, undefined, {
    name: SUPPORT_ALBY_CONNECTION_NAME,
  });
  const { data: albyMe } = useAlbyMe();
  const { order } = useChannelOrderStore();
  const location = useLocation();
  const { toast } = useToast();

  const [hiddenUntil, setHiddenUntil] = React.useState(
    localStorage.getItem(localStorageKeys.supportAlbySidebarHintHiddenUntil)
  );

  // User has a channel order
  if (
    !location.pathname.startsWith("/channels/") &&
    order &&
    order.status !== "pay"
  ) {
    return (
      <SidebarHintCard
        icon={ZapIcon}
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
        icon={ListTodoIcon}
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

  const showSupport =
    supportAlbyAppsData &&
    supportAlbyAppsData.apps.length === 0 &&
    !albyMe?.subscription.plan_code;

  if (
    !location.pathname.startsWith("/support-alby") &&
    showSupport &&
    (!hiddenUntil || new Date() >= new Date(hiddenUntil))
  ) {
    return (
      <SidebarHintCard
        onClose={() => {
          // Set the date to the next 21st of the month
          const now = new Date();
          const next21st = new Date(
            now.getFullYear(),
            now.getMonth() + (now.getDate() >= 21 ? 1 : 0),
            21
          ).toString();
          localStorage.setItem(
            localStorageKeys.supportAlbySidebarHintHiddenUntil,
            next21st
          );
          setHiddenUntil(next21st);
          toast({ title: "No worries, we'll remind you again!" });
        }}
        icon={HeartIcon}
        title="Support Alby Hub"
        description="See how you can support the development of Alby Hub"
        buttonText="Become a Supporter"
        buttonLink="/support-alby"
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
  onClose?: () => void;
};
function SidebarHintCard({
  title,
  description,
  icon: Icon,
  buttonText,
  buttonLink,
  onClose,
}: SidebarHintCardProps) {
  return (
    <Card>
      <CardHeader>
        <Icon className="h-8 w-8 mb-2" />
        <CardTitle>{title}</CardTitle>
        {onClose && (
          <button
            className="absolute top-4 right-4 text-muted-foreground hover:text-primary"
            onClick={onClose}
          >
            <XIcon name="X" />
          </button>
        )}
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        <div className="text-muted-foreground text-sm">{description}</div>
        <Link to={buttonLink}>
          <Button size="sm" className="w-full">
            {buttonText}
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
}

export default SidebarHint;
