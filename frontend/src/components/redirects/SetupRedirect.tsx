import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";
import { SetupCode } from "src/screens/setup/SetupCode";
import useSetupStore from "src/state/SetupStore";

export function SetupRedirect() {
  const { data: info } = useInfo();
  const setupCode = useSetupStore((store) => store.setupCode);
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info) {
      return;
    }
    if (info.setupCompleted && info.running) {
      useSetupStore.getState().reset();
      navigate("/");
      return;
    }
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  if (info.setupCodeRequired && !setupCode) {
    return <SetupCode />;
  }

  return <Outlet />;
}
