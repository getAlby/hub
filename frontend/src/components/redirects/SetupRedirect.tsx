import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
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
    // if (info.setupCompleted) {
    //   navigate("/");
    // }
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return <Outlet />;
}
