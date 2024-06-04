import { Link2, LucideIcon, Plane, ShieldAlert, Zap } from "lucide-react";
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
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { backendTypeHasMnemonic } from "src/lib/utils";
import useChannelOrderStore from "src/state/ChannelOrderStore";

function SidebarHint() {
  const { data: channels } = useChannels();
  const { data: albyBalance } = useAlbyBalance();
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();
  const { order } = useChannelOrderStore();
  const location = useLocation();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();

  // Don't distract with hints while opening a channel
  if (
    location.pathname.endsWith("/channels/order") ||
    location.pathname.endsWith("/channels/new")
  ) {
    return null;
  }

  // User has a channel order
  if (order && order.status !== "pay") {
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
    albyBalance.sats * (1 - ALBY_SERVICE_FEE) >
      ALBY_MIN_BALANCE + 50000 /* accomodate for onchain fees */
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
  if (
    (info?.backendType === "LDK" || info?.backendType === "GREENLIGHT") &&
    channels?.length === 0
  ) {
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

  // User has not linked their hub to their Alby Account
  if (
    albyMe &&
    nodeConnectionInfo &&
    albyMe?.keysend_pubkey !== nodeConnectionInfo?.pubkey
  ) {
    return (
      <SidebarHintCard
        icon={Link2}
        title="Link your Hub"
        description="Finish the setup by linking your Alby Account to this hub."
        buttonText="Link Hub"
        buttonLink="/settings"
      />
    );
  }

  if (
    info &&
    backendTypeHasMnemonic(info.backendType) &&
    (!info.nextBackupReminder ||
      new Date(info.nextBackupReminder).getTime() < new Date().getTime())
  ) {
    return (
      <SidebarHintCard
        icon={ShieldAlert}
        title="Backup Your Keys"
        description=" Not backing up your key might result in permanently losing
              access to your funds."
        buttonText="Backup Now"
        buttonLink="/settings/key-backup"
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
