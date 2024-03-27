import { useInfo } from "src/hooks/useInfo";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";

export function ChannelsRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info || (info.running && info.unlocked)) {
      return;
    }
    navigate("/");
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return <Outlet />;
}
