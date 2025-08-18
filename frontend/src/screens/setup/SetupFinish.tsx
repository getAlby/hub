import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import LottieLoading from "src/components/LottieLoading";
import { Button } from "src/components/ui/button";
import { ToastSignature, useToast } from "src/components/ui/use-toast";

import { useInfo } from "src/hooks/useInfo";
import { saveAuthToken } from "src/lib/auth";
import useSetupStore from "src/state/SetupStore";
import { AuthTokenResponse, SetupNodeInfo } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

let lastStartupErrorTime: string;
export function SetupFinish() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: info } = useInfo(true); // poll the info endpoint to auto-redirect when app is running

  const [loading, setLoading] = React.useState(false);
  const [connectionError, setConnectionError] = React.useState(false);
  const hasFetchedRef = React.useRef(false);

  const startupError = info?.startupError;
  const startupErrorTime = info?.startupErrorTime;

  React.useEffect(() => {
    // lastStartupErrorTime check is required because user may leave page and come back
    // after re-configuring settings
    if (
      startupError &&
      startupErrorTime &&
      startupErrorTime !== lastStartupErrorTime
    ) {
      lastStartupErrorTime = startupErrorTime;
      toast({
        title: "Failed to start",
        description: startupError,
        variant: "destructive",
      });
      setLoading(false);
      setConnectionError(true);
    }
  }, [startupError, toast, startupErrorTime]);

  useEffect(() => {
    if (!loading) {
      return;
    }
    const timer = setTimeout(() => {
      // SetupRedirect takes care of redirection once info.running is true
      // if it still didn't redirect after 30 seconds, we show an error
      // Typically initial startup should complete in less than 10 seconds.
      setLoading(false);
      setConnectionError(true);
    }, 30000);

    return () => {
      clearTimeout(timer);
    };
  }, [loading]);

  useEffect(() => {
    if (!info) {
      return;
    }
    // ensure setup call is only called once
    if (hasFetchedRef.current) {
      return;
    }
    hasFetchedRef.current = true;

    (async () => {
      setLoading(true);
      const succeeded = await finishSetup(
        useSetupStore.getState().nodeInfo,
        useSetupStore.getState().unlockPassword,
        toast
      );
      // only setup call is successful as start is async
      if (!succeeded) {
        setLoading(false);
        setConnectionError(true);
      }
    })();
  }, [navigate, toast, info]);

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
        <LottieLoading size={400} />
        <h1 className="font-semibold text-lg font-headline">
          Setting up your Hub...
        </h1>
      </div>
    </Container>
  );
}

const finishSetup = async (
  nodeInfo: SetupNodeInfo,
  unlockPassword: string,
  toast: ToastSignature
): Promise<boolean> => {
  try {
    await request("/api/setup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...nodeInfo,
        unlockPassword,
      }),
    });

    const authTokenResponse = await request<AuthTokenResponse>("/api/start", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        unlockPassword,
      }),
    });
    if (authTokenResponse) {
      saveAuthToken(authTokenResponse.token);
    }
    return true;
  } catch (error) {
    handleRequestError(toast, "Failed to connect", error);
    return false;
  }
};
