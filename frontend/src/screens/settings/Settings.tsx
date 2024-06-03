import { CircleCheck, Link2Off } from "lucide-react";
import { useEffect, useState } from "react";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { toast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCSRF } from "src/hooks/useCSRF";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { request } from "src/utils/request";

function Settings() {
  const { data: csrf } = useCSRF();
  const { data: me } = useAlbyMe();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const [loading, setLoading] = useState(false);
  const [loadingInfo, setLoadingInfo] = useState(true);
  const [linked, setLinked] = useState(false);

  useEffect(() => {
    if (me && nodeConnectionInfo) {
      setLinked(me?.keysend_pubkey === nodeConnectionInfo.pubkey);
      setLoadingInfo(false);
    }
  }, [me, nodeConnectionInfo]);

  async function linkAccount() {
    try {
      setLoading(true);
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      await request("/api/alby/link-account", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      });
      setLinked(true);
      toast({
        title:
          "Your Alby Hub has successfully been linked to your Alby Account",
      });
    } catch (e) {
      toast({
        title: "Your Alby Hub couldn't be linked to your Alby Account",
        description: "Did you already link another Alby Hub?",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <SettingsHeader
        title="General"
        description="Adjust general settings of your Alby Hub"
      />
      <div>
        <Card>
          <CardHeader>
            <CardTitle>Alby Account</CardTitle>
            <CardDescription>
              Link your lightning address & other apps to this hub.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground text-sm">Status</div>
            {loadingInfo && <Loading />}
            {!loadingInfo && (
              <div className="flex flex-row gap-2 items-center">
                {linked && (
                  <>
                    <CircleCheck className="w-4 h-4" />
                    <p className="font-medium">Linked</p>
                  </>
                )}
                {!linked && me?.shared_node && (
                  <>
                    <Link2Off className="w-4 h-4" />
                    <p className="font-medium">Not Linked</p>
                  </>
                )}
                {!linked && !me?.shared_node && (
                  <>
                    <Link2Off className="w-4 h-4" />
                    <p className="font-medium">Linked to a different wallet</p>
                  </>
                )}
              </div>
            )}
          </CardContent>
          {!loadingInfo && !linked && me?.shared_node && (
            <CardFooter>
              <LoadingButton loading={loading} onClick={linkAccount}>
                Link now
              </LoadingButton>
            </CardFooter>
          )}
        </Card>
      </div>
    </>
  );
}

export default Settings;
