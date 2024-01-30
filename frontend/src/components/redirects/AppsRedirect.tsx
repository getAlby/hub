import { useInfo } from "src/hooks/useInfo";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";

export function AppsRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  // TODO: re-add login redirect: https://github.com/getAlby/nostr-wallet-connect/commit/59b041886098dda4ff38191e3dd704ec36360673
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
