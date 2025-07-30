import { useState } from "react";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";

import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { BudgetRenewalType, ListAppsResponse } from "src/types";
import { request } from "src/utils/request";
import { KeyedMutator } from "swr";

export enum LinkStatus {
  SharedNode,
  ThisNode,
  OtherNode,
  Unlinked,
}

export function useLinkAccount(
  reloadAlbyAccountApp: KeyedMutator<ListAppsResponse>
) {
  const { data: me, mutate: reloadAlbyMe } = useAlbyMe();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  let linkStatus: LinkStatus | undefined;
  if (me && nodeConnectionInfo) {
    if (me.keysend_pubkey === nodeConnectionInfo.pubkey) {
      linkStatus = LinkStatus.ThisNode;
    } else if (me.shared_node) {
      linkStatus = LinkStatus.SharedNode;
    } else if (me.keysend_pubkey) {
      linkStatus = LinkStatus.OtherNode;
    } else {
      linkStatus = LinkStatus.Unlinked;
    }
  }

  const loadingLinkStatus = linkStatus === undefined;

  async function linkAccount(budget: number, renewal: BudgetRenewalType) {
    try {
      setLoading(true);

      await request("/api/alby/link-account", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          budget,
          renewal,
        }),
      });
      // update the link status and get the newly-created Alby Account app
      await Promise.all([reloadAlbyMe(), reloadAlbyAccountApp()]);
      toast({
        title:
          "Your Alby Hub has successfully been linked to your Alby Account",
      });
    } catch (e) {
      console.error(e);
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
