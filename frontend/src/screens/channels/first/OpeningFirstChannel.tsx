import React from "react";
import { useNavigate } from "react-router-dom";
import { ChannelWaitingForConfirmations } from "src/components/channels/ChannelWaitingForConfirmations";
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

  return <ChannelWaitingForConfirmations channel={firstChannel} />;
}
