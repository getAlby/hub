import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

let didSetupThisSession = false;
export function SetupRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info) {
      return;
    }
    if (didSetupThisSession) {
      // ensure redirect does not happen as node may still be starting
      // which would then incorrectly redirect to the login page
      console.info("Skipping setup redirect on initial setup");
      return;
    }
    if (info.setupCompleted) {
      navigate("/");
      return;
    }
    didSetupThisSession = true;
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return <Outlet />;
}
