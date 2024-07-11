import React, { useEffect } from "react";
import Lottie from "react-lottie";
import { useNavigate } from "react-router-dom";
import animationData from "src/assets/lotties/loading.json";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import { toast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";
import { SetupNodeInfo } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function SetupFinish() {
  const navigate = useNavigate();
  const { nodeInfo, unlockPassword } = useSetupStore();
  useInfo(true); // poll the info endpoint to auto-redirect when app is running
  const { data: csrf } = useCSRF();
  const [loading, setLoading] = React.useState(false);
  const [connectionError, setConnectionError] = React.useState(false);
  const hasFetchedRef = React.useRef(false);

  const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice",
    },
  };

  useEffect(() => {
    if (!loading) {
      return;
    }
    const timer = setTimeout(() => {
      // SetupRedirect takes care of redirection once info.running is true
      // if it still didn't redirect after 3 minutes, we show an error
      setLoading(false);
      setConnectionError(true);
    }, 180000);

    return () => {
      clearTimeout(timer);
    };
  }, [loading]);

  useEffect(() => {
    // ensure setup call is only called once
    if (!csrf || hasFetchedRef.current) {
      return;
    }
    hasFetchedRef.current = true;

    (async () => {
      setLoading(true);
      const succeeded = await finishSetup(csrf, nodeInfo, unlockPassword);
      // only setup call is successful as start is async
      if (!succeeded) {
        setLoading(false);
        setConnectionError(true);
      }
    })();
  }, [csrf, nodeInfo, navigate, unlockPassword]);

  if (connectionError) {
    return (
      <Container>
        <div className="flex flex-col gap-5 text-center items-center">
          <div className="grid gap-2">
            <h1 className="font-semibold text-lg">Connection Failed</h1>
            <p>Please check your node configuration and try again.</p>
          </div>
          <Button
            onClick={() => {
              navigate(-1);
            }}
          >
            Try again
          </Button>
        </div>
      </Container>
    );
  }

  return (
    <Container>
      <div className="flex flex-col gap-5 justify-center text-center">
        <Lottie options={defaultOptions} height={400} width={400} />
        <h1 className="font-semibold text-lg font-headline">
          Setting up your Hub...
        </h1>
      </div>
    </Container>
  );
}

const finishSetup = async (
  csrf: string,
  nodeInfo: SetupNodeInfo,
  unlockPassword: string
): Promise<boolean> => {
  try {
    await request("/api/setup", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...nodeInfo,
        unlockPassword,
      }),
    });
    await request("/api/start", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        unlockPassword,
      }),
    });
    return true;
  } catch (error) {
    handleRequestError(toast, "Failed to connect", error);
    return false;
  }
};
