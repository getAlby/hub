import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ChannelWaitingForConfirmations } from "src/components/channels/ChannelWaitingForConfirmations";
import { useChannels } from "src/hooks/useChannels";
import { useSyncWallet } from "src/hooks/useSyncWallet";

export function OpeningAutoChannel() {
  useSyncWallet();
  const { data: channels } = useChannels(true);
  const navigate = useNavigate();

  const { state } = useLocation();
  const newChannelId = state?.newChannelId as string | undefined;

  const channel = channels?.find(
    (channel) => channel.id && channel.id === newChannelId
  );

  React.useEffect(() => {
    if (channel?.active) {
      navigate("/channels/auto/opened");
    }
  }, [channel, navigate]);

  return <ChannelWaitingForConfirmations channel={channel} />;
}
