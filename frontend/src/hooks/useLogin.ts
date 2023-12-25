import { useLocation, useNavigate } from "react-router-dom";
import { useInfo } from "./useInfo";
import React from "react";

export function useLogin() {
  // TODO:
  const { data: info, isLoading } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    // TODO: Use the location to redirect back in /alby/auth?c=
    if (!isLoading && !info?.user && location.pathname !== "/login") {
      navigate("/login");
    }
  }, [info?.user, isLoading, location.pathname, navigate]);

  return null;
}
