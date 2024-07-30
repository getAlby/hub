import { Footprints } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import EmptyState from "src/components/EmptyState";
import Loading from "src/components/Loading";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useChannels } from "src/hooks/useChannels";
import { useSyncWallet } from "src/hooks/useSyncWallet";

export function OpeningFirstChannel() {
  useSyncWallet();
  const { data: channels } = useChannels(true);
  const navigate = useNavigate();

  const firstChannel = channels?.[0];

  React.useEffect(() => {
    if (firstChannel?.active) {
      navigate("/channels/first/opened");
    }
  }, [firstChannel, navigate]);

  if (!firstChannel || !firstChannel.confirmationsRequired) {
    // 0-conf channel, this should only take a few seconds
    return <Loading />;
  }

  return (
    <>
      <div className="flex flex-col justify-center gap-2">
        <Card>
          <CardHeader>
            <CardTitle>Your channel is being opened</CardTitle>
            <CardDescription>
              Waiting for {firstChannel.confirmationsRequired} confirmations
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-row gap-2">
              <Loading />
              {firstChannel.confirmations ?? "0"} /{" "}
              {firstChannel.confirmationsRequired ?? "unknown"} confirmations
            </div>
          </CardContent>
        </Card>
        <div className="w-full mt-40 flex flex-col items-center justify-center">
          <EmptyState
            icon={Footprints}
            title="Browse While You Wait"
            description="Feel free to leave this page or browse around Alby Hub! We'll send you an email as soon as your channel is active."
            buttonText="Explore Apps"
            buttonLink="/appstore"
          />
        </div>
      </div>
    </>
  );
}
