import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import Loading from "src/components/Loading";
import { NodeUnavailable } from "src/components/NodeUnavailable";
import { localStorageKeys } from "src/constants";
import { useInfo } from "src/hooks/useInfo";
import { useNodeStatus } from "src/hooks/useNodeStatus";

const nodeDependentPaths = ["/home", "/wallet", "/channels", "/peers"];

export function DefaultRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();
  const canAccessApp =
    !!info?.running &&
    !!info.unlocked &&
    (info.albyAccountConnected || !info.albyUserIdentifier);
  const isNodeDependentPath = nodeDependentPaths.some(
    (path) =>
      location.pathname === path || location.pathname.startsWith(`${path}/`)
  );
  const {
    data: nodeStatus,
    error: nodeStatusError,
    isLoading: isNodeStatusLoading,
    mutate: refetchNodeStatus,
  } = useNodeStatus(canAccessApp && isNodeDependentPath);

  React.useEffect(() => {
    if (!info || canAccessApp) {
      return;
    }
    const returnTo = location.pathname + location.search;
    window.localStorage.setItem(localStorageKeys.returnTo, returnTo);
    navigate("/");
  }, [canAccessApp, info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  if (canAccessApp && isNodeDependentPath && isNodeStatusLoading) {
    return <Loading />;
  }

  if (
    canAccessApp &&
    isNodeDependentPath &&
    !isNodeStatusLoading &&
    (nodeStatusError || !nodeStatus?.isReady)
  ) {
    return <NodeUnavailable onRetry={() => void refetchNodeStatus()} />;
  }

  return <Outlet />;
}
