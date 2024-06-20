import { useState } from "react";
import { toast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useCSRF } from "src/hooks/useCSRF";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { request } from "src/utils/request";

export enum LinkStatus {
  SharedNode,
  ThisNode,
  OtherNode,
}

export function useLinkAccount() {
  const { data: csrf } = useCSRF();
  const { data: me, mutate: reloadAlbyMe } = useAlbyMe();
  const { mutate: reloadApps } = useApps();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const [loading, setLoading] = useState(false);

  let linkStatus: LinkStatus | undefined;
  if (me && nodeConnectionInfo) {
    if (me?.keysend_pubkey === nodeConnectionInfo.pubkey) {
      linkStatus = LinkStatus.ThisNode;
    } else if (me.shared_node) {
      linkStatus = LinkStatus.SharedNode;
    } else {
      linkStatus = LinkStatus.OtherNode;
    }
  }

  const loadingLinkStatus = linkStatus === undefined;

  async function linkAccount() {
    try {
      setLoading(true);
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      await request("/api/alby/link-account", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
      });
      // update the link status and get the newly-created Alby Account app
      await Promise.all([reloadAlbyMe(), reloadApps()]);
      toast({
        title:
          "Your Alby Hub has successfully been linked to your Alby Account",
      });
    } catch (e) {
      toast({
        title: "Your Alby Hub couldn't be linked to your Alby Account",
        description: "Did you already link another Alby Hub?",
      });
    } finally {
      setLoading(false);
    }
  }

  return { loading, loadingLinkStatus, linkStatus, linkAccount };
}
