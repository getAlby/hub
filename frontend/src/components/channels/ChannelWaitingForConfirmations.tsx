import Lottie from "react-lottie";
import animationData from "src/assets/lotties/loading.json";
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

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  return (
    <div className="flex flex-col gap-4">
      <p className="text-muted-foreground ">
        You can now leave this page.We’ll notify your by email once your channel
        is active.
      </p>
      <div className="w-80 flex justify-center">
        <Card className="text-center">
          <CardHeader>
            <CardTitle>Opening new lightning channel...</CardTitle>
            <CardDescription>
              Waiting for {channel.confirmationsRequired} confirmations
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-5 justify-center text-center">
              <Lottie options={defaultOptions} height={256} width={256} />
              {channel.confirmations ?? "0"} /{" "}
              {channel.confirmationsRequired ?? "unknown"} confirmations
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
