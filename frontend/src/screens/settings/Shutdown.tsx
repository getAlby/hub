import { Power } from "lucide-react";
import React, { useState } from "react";
import Lottie from "react-lottie";
import { useNavigate } from "react-router-dom";
import animationData from "src/assets/lotties/loading.json";
import SettingsHeader from "src/components/SettingsHeader";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

function Shutdown() {
  const [shuttingDown, setShuttingDown] = useState(false);
  const { mutate: refetchInfo } = useInfo();
  const navigate = useNavigate();
  const { toast } = useToast();

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  const shutdown = React.useCallback(async () => {
    setShuttingDown(true);
    try {
      await request("/api/stop", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      await refetchInfo();
      setShuttingDown(false);
      navigate("/", { replace: true });
      toast({ title: "Your node has been turned off." });
    } catch (error) {
      console.error(error);
      toast({
        title: "Failed to shutdown node: " + error,
        variant: "destructive",
      });
    }
  }, [navigate, refetchInfo, toast]);

  return (
    <>
      <SettingsHeader
        title="Shutdown"
        description="Shutting down Hub temporarily disables all connections and transacting with your wallet until you turn on your Hub with your unlock password."
      />
      {shuttingDown ? (
        <div className="flex flex-col gap-5 justify-center text-center">
          <Lottie options={defaultOptions} height={400} width={400} />
          <h1 className="font-semibold text-lg font-headline">
            Shutting down...
          </h1>
        </div>
      ) : (
        <>
          <form onSubmit={shutdown} className="max-w-md flex flex-col gap-8">
            <div className="flex justify-start">
              <LoadingButton variant="destructive" loading={shuttingDown}>
                <div className="flex gap-2 items-center">
                  <Power className="w-4 h-4" /> Shutdown
                </div>
              </LoadingButton>
            </div>
          </form>
        </>
      )}
    </>
  );
}

export default Shutdown;
