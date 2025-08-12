import { ArrowLeftIcon, ExternalLinkIcon, HandCoinsIcon } from "lucide-react";
import { useLocation } from "react-router-dom";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { ExternalLinkButton, LinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

import TickSVG from "public/images/illustrations/tick.svg";
import Lottie from "react-lottie";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { useTheme } from "src/components/ui/theme-provider";
import { useInfo } from "src/hooks/useInfo";
import { useMempoolApi } from "src/hooks/useMempoolApi";
import { MempoolTransaction } from "src/types";

export default function OnchainSuccess() {
  const { state } = useLocation();
  const { data: info } = useInfo();
  const { isDarkMode } = useTheme();

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: isDarkMode ? animationDataDark : animationDataLight,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  const amount = state?.amount as number;
  const txId = state?.txId as string;

  const { data: mempoolTx } = useMempoolApi<MempoolTransaction>(
    `/tx/${txId}`,
    3000
  );

  if (!info) {
    return <Loading />;
  }

  return (
    <div className="grid gap-4">
      <AppHeader title="Send to On-chain" />
      <div className="w-full md:max-w-lg">
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="text-center">
              {mempoolTx?.status.confirmed
                ? "Payment Successful"
                : "Awaiting Confirmation"}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6">
            {mempoolTx?.status.confirmed ? (
              <img src={TickSVG} className="w-48" />
            ) : (
              <Lottie options={defaultOptions} height={288} width={288} />
            )}
            <div className="flex flex-col gap-1 items-center">
              <p className="text-2xl font-medium slashed-zero">
                {new Intl.NumberFormat().format(amount)} sats
              </p>
              <FormattedFiatAmount amount={amount} className="text-xl" />
            </div>
            {!mempoolTx?.status.confirmed && (
              <p className="text-muted-foreground text-center">
                You can now leave this page.
                <br />
                You'll be notified when transaction arrives.
              </p>
            )}
          </CardContent>
          <CardFooter className="flex flex-col gap-2 pt-2">
            <ExternalLinkButton
              to={`${info?.mempoolUrl}/tx/${txId}`}
              variant="outline"
              className="w-full"
            >
              <ExternalLinkIcon className="w-4 h-4 mr-2" />
              View on Mempool
            </ExternalLinkButton>
            <LinkButton to="/wallet/send" variant="outline" className="w-full">
              <HandCoinsIcon className="w-4 h-4 mr-2" />
              Make Another Payment
            </LinkButton>
            <LinkButton to="/wallet" variant="link" className="w-full">
              <ArrowLeftIcon className="w-4 h-4 mr-2" />
              Back to Wallet
            </LinkButton>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
