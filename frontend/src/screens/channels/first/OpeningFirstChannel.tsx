import React from "react";
import { useNavigate } from "react-router";
import { ChannelWaitingForConfirmations } from "src/components/channels/ChannelWaitingForConfirmations";
import { useChannels } from "src/hooks/useChannels";
import { useSyncWallet } from "src/hooks/useSyncWallet";

type OpeningFirstChannelProps = {
  openedPath?: string;
};

export function OpeningFirstChannel({
  openedPath = "/channels/first/opened",
}: OpeningFirstChannelProps) {
  useSyncWallet();
  const { data: channels } = useChannels(true);
  const navigate = useNavigate();

  const firstChannel = channels?.[0];

  React.useEffect(() => {
    if (firstChannel?.active) {
      navigate(openedPath);
    }
  }, [firstChannel, navigate, openedPath]);

  return <ChannelWaitingForConfirmations channel={firstChannel} />;
}
