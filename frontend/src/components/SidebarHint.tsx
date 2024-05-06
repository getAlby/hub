import { LucideIcon, Plane, ShieldAlert, Zap } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ALBY_MIN_BALANCE, ALBY_SERVICE_FEE } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import useChannelOrderStore from "src/state/ChannelOrderStore";

function SidebarHint() {
  const { data: channels } = useChannels();
  const { data: albyBalance } = useAlbyBalance();
  const { data: info } = useInfo();
  const { order } = useChannelOrderStore();
  const location = useLocation();

  // Don't distract with hints while opening a channel
  if (
    location.pathname.endsWith("/channels/order") ||
    location.pathname.endsWith("/channels/new")
  ) {
    return null;
  }

  // User has a channel order
  if (order) {
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

  // User has funds to migrate
  if (
    info?.backendType === "LDK" &&
    albyBalance &&
    albyBalance.sats * (1 - ALBY_SERVICE_FEE) > ALBY_MIN_BALANCE
  ) {
    return (
      <SidebarHintCard
        icon={Plane}
        title="Migrate Alby Funds"
        description="You have funds on your Alby shared account ready to migrate to your new node"
        buttonText="Migrate Now"
        buttonLink="/onboarding/lightning/migrate-alby"
      />
    );
  }

  // User has no channels yet
  if (channels?.length === 0) {
    return (
      <SidebarHintCard
        icon={Zap}
        title="Open Your First Channel"
        description="Deposit bitcoin by onchain or lightning payment to start using your new wallet."
        buttonText="Begin Now"
        buttonLink="/channels/new"
      />
    );
  }

  if (info?.showBackupReminder) {
    return (
      <SidebarHintCard
        icon={ShieldAlert}
        title="Backup Your Keys"
        description=" Not backing up your key might result in permanently losing
              access to your funds."
        buttonText="Backup Now"
        buttonLink="/settings/backup"
      />
    );
  }
}

type SidebarHintCardProps = {
  title: string;
  description: string;
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
