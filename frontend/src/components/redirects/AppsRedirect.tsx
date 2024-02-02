import { useInfo } from "src/hooks/useInfo";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";
import { localStorageKeys } from "src/constants";

export function AppsRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info || (info.running && info.unlocked)) {
      window.localStorage.removeItem(localStorageKeys.returnTo);
      return;
    }
    const returnTo = location.pathname + location.search;
    window.localStorage.setItem(localStorageKeys.returnTo, returnTo);
    navigate("/");
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return <Outlet />;
}
