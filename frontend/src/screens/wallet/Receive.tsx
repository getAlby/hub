import { LinkIcon, ZapIcon } from "lucide-react";
import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { ReceiveToOnchain } from "src/components/ReceiveToOnchain";
import { ReceiveToSpending } from "src/components/ReceiveToSpending";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "src/components/ui/tabs";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export default function Receive() {
  const { data: info } = useInfo();
  const { data: me, error: meError } = useAlbyMe();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const activeTab =
    searchParams.get("type") === "onchain" ? "onchain" : "spending";

  const isSpendingTab = activeTab === "spending";
  const isLightningAddressLoading =
    isSpendingTab && info?.albyAccountConnected && !me && !meError;

  // TODO: remove this once we have a CTA to connect an Alby Account to use a lightning address
  useEffect(() => {
    if (!isSpendingTab || !info || isLightningAddressLoading) {
      return;
    }
    if (!info.albyAccountConnected) {
      navigate("/wallet/receive/invoice", { replace: true });
      return;
    }
    if (meError) {
      toast.error("Failed to load lightning address");
      navigate("/wallet/receive/invoice", { replace: true });
    }
  }, [info, isLightningAddressLoading, isSpendingTab, meError, navigate]);

  if (!info || isLightningAddressLoading) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader pageTitle="Receive" title="Receive" />
      <div className="w-full max-w-lg">
        <Tabs
          value={activeTab}
          onValueChange={() => {
            setSearchParams(isSpendingTab ? { type: "onchain" } : {}, {
              replace: true,
            });
          }}
          className="w-full"
        >
          <TabsList className="w-full mb-2">
            <TabsTrigger
              value="spending"
              className="flex gap-2 items-center w-full"
            >
              <ZapIcon className="size-4" />
              To Spending Balance
            </TabsTrigger>
            <TabsTrigger
              value="onchain"
              className="flex gap-2 items-center w-full"
            >
              <LinkIcon className="size-4" />
              To On-chain Balance
            </TabsTrigger>
          </TabsList>
          <TabsContent value="spending">
            <ReceiveToSpending />
          </TabsContent>
          <TabsContent value="onchain">
            <ReceiveToOnchain />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
