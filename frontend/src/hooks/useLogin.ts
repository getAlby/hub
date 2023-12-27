import { useLocation, useNavigate } from "react-router-dom";
import { useInfo } from "./useInfo";
import React from "react";

const RETURN_TO_KEY = "returnTo";

export function useLogin() {
  const { data: info, isLoading } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!isLoading) {
      if (info?.user) {
        const returnTo = window.localStorage.getItem(RETURN_TO_KEY);
        if (returnTo) {
          window.localStorage.removeItem(RETURN_TO_KEY);
          navigate(returnTo);
        }
      }
      if (!info?.user && ["/login", "/about"].indexOf(location.pathname) < 0) {
        const returnTo = location.pathname + location.search;
        window.localStorage.setItem(RETURN_TO_KEY, returnTo);

        navigate("/login");
      }
    }
  }, [info?.user, isLoading, location.pathname, location.search, navigate]);

  return null;
}
