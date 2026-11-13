import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";

export function SetupRedirect() {
  const { data: info } = useInfo();
  const setupCode = useSetupStore((store) => store.setupCode);
  const location = useLocation();
  const navigate = useNavigate();

  const isSetupCodeMissing = !!info?.setupCodeRequired && !setupCode;

  React.useEffect(() => {
    if (!info) {
      return;
    }
    if (info.setupCompleted && info.running) {
      useSetupStore.getState().reset();
      navigate("/");
      return;
    }
    if (isSetupCodeMissing) {
      // the setup code is entered on the welcome page
      navigate("/welcome", { replace: true });
      return;
    }
  }, [info, isSetupCodeMissing, location, navigate]);

  if (!info || isSetupCodeMissing) {
    return <Loading />;
  }

  return <Outlet />;
}
