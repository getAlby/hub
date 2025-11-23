import { FootprintsIcon } from "lucide-react";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Channel } from "src/types";

export function ChannelWaitingForConfirmations({
  channel,
}: {
  channel: Channel | undefined;
}) {
  if (!channel?.confirmationsRequired) {
    // 0-conf channel or waiting for transaction to be broadcasted, this should only take a few seconds
    return <Loading />;
  }

  return (
    <div className="flex flex-col justify-center gap-2">
      <Card>
        <CardHeader>
          <CardTitle>Your channel is being opened</CardTitle>
          <CardDescription>
            Waiting for {channel.confirmationsRequired} confirmations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-row gap-2">
            <Loading />
            {channel.confirmations ?? "0"} /{" "}
            {channel.confirmationsRequired ?? "unknown"} confirmations
          </div>
        </CardContent>
      </Card>
      <div className="w-full mt-40 flex flex-col items-center justify-center">
        <EmptyState
          icon={FootprintsIcon}
          title="Browse While You Wait"
          description="Feel free to leave this page or browse around Alby Hub! We'll send you an email as soon as your channel is active."
          buttonText="Explore Apps"
          buttonLink="/apps?tab=app-store"
        />
      </div>
    </div>
  );
}
