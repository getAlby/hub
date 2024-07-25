import React from "react";
import { useNavigate } from "react-router-dom";
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

  if (!firstChannel) {
    return <Loading />;
  }

  return (
    <>
      <div className="flex flex-col justify-center gap-2">
        <Card>
          <CardHeader>
            <CardTitle>Your channel is being opened</CardTitle>
            <CardDescription>
              Waiting for {firstChannel?.confirmationsRequired ?? "unknown"}{" "}
              confirmations
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-row gap-2">
              <Loading />
              {firstChannel?.confirmations ?? "0"} /{" "}
              {firstChannel?.confirmationsRequired ?? "unknown"} confirmations
            </div>
          </CardContent>
        </Card>
        <div className="w-full mt-40 gap-20 flex flex-col items-center justify-center">
          <p>Feel free to leave this page or browse around Alby Hub!</p>
          <p>We'll send you an email as soon as your channel is active.</p>
        </div>
      </div>
    </>
  );
}
