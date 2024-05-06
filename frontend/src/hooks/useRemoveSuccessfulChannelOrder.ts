import React from "react";
import { useLocation } from "react-router-dom";
import useChannelOrderStore from "src/state/ChannelOrderStore";

export function useRemoveSuccessfulChannelOrder() {
  const location = useLocation();
  const [prevLocation, setPrevLocation] = React.useState(location.pathname);

  React.useEffect(() => {
    if (
      location.pathname != "/channels/order" &&
      prevLocation === "/channels/order" &&
      useChannelOrderStore.getState().order?.status === "success"
    ) {
      useChannelOrderStore.getState().removeOrder();
    }
    setPrevLocation(location.pathname);
  }, [location.pathname, prevLocation]);
}
