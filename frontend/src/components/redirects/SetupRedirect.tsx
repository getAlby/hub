import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

export function SetupRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info) {
      return;
    }
    if (info.setupCompleted && info.running) {
      if (location.pathname.startsWith("/setup/first-channel")) {
        return;
      }
      if (location.pathname.startsWith("/setup/finish")) {
        navigate("/setup/first-channel", { replace: true });
        return;
      }
      navigate("/");
      return;
    }
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return <Outlet />;
}
