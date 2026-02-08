import { useCallback, useEffect, useRef, useState } from "react";
import { Channel } from "src/types";
import { request } from "src/utils/request";
import {
  fetchNodeAlias,
  getHopColor,
  MAX_HOPS,
  MAX_NODES,
  processGossipChannels,
} from "./graphUtils";
import { GraphLink, GraphNode, NetworkGraphApiResponse } from "./types";

export function useNetworkGraph(
  ourPubkey: string | undefined,
  channels: Channel[] | undefined
) {
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [links, setLinks] = useState<GraphLink[]>([]);
  const [loading, setLoading] = useState(false);
  const abortRef = useRef(false);
  const hasRunRef = useRef(false);
  // Use a ref for channels so the expand callback doesn't depend on it.
  // This prevents SWR revalidation from recreating expand mid-flight,
  // which would trigger the abort cleanup and kill hop 3+ expansion.
  const channelsRef = useRef(channels);
  channelsRef.current = channels;

  const expand = useCallback(async () => {
    const ch = channelsRef.current;
    if (!ourPubkey || !ch) {
      return;
    }
    // Prevent re-running when channels reference changes (SWR revalidation)
    if (hasRunRef.current) {
      return;
    }
    hasRunRef.current = true;

    abortRef.current = false;
    setLoading(true);

    const directPeerPubkeys = new Set(ch.map((c) => c.remotePubkey));

    const knownNodeIds = new Set<string>();
    const knownChannelIds = new Set<string>();
    let currentNodes: GraphNode[] = [];
    let currentLinks: GraphLink[] = [];
    // Collect aliases separately so we can apply them in one pass at the end,
    // instead of scanning the entire nodes array per alias update.
    const aliasMap = new Map<string, string>();

    // Hop 0: Our node
    knownNodeIds.add(ourPubkey);
    currentNodes.push({
      id: ourPubkey,
      alias: "You",
      hop: 0,
      isOurNode: true,
      hasChannel: true,
      color: getHopColor(0),
    });

    // Hop 1: Add direct channel partners from local channel data.
    // This is essential because private channels don't appear in gossip.
    const hop1PeerIds: string[] = [];
    for (const channel of ch) {
      const peerId = channel.remotePubkey;
      if (!knownNodeIds.has(peerId)) {
        knownNodeIds.add(peerId);
        hop1PeerIds.push(peerId);
        currentNodes.push({
          id: peerId,
          alias: peerId.slice(0, 8),
          hop: 1,
          isOurNode: false,
          hasChannel: true,
          color: getHopColor(1),
        });
      }

      const channelId = channel.id;
      if (!knownChannelIds.has(channelId)) {
        knownChannelIds.add(channelId);
        currentLinks.push({
          source: ourPubkey,
          target: peerId,
          capacity: Math.round(
            (channel.localBalance + channel.remoteBalance) / 1000
          ),
          isOurChannel: true,
        });
      }
    }

    // Fetch gossip for hop 1 peers: resolve aliases AND extract hop 2 channels
    let hop2Frontier = new Set<string>();
    if (hop1PeerIds.length > 0) {
      try {
        const response = await request<NetworkGraphApiResponse>(
          `/api/node/network-graph?nodeIds=${hop1PeerIds.join(",")}`
        );
        if (response) {
          console.info(
            `[NetworkGraph] Gossip for hop 1 peers: ${response.nodes.length} nodes, ${response.channels.length} channels`
          );

          for (const graphNode of response.nodes) {
            const nodeAlias = graphNode.node?.alias || graphNode.node?.Alias;
            if (nodeAlias) {
              aliasMap.set(graphNode.nodeId, nodeAlias);
            }
          }

          if (response.channels.length > 0) {
            const result = processGossipChannels(
              response.channels,
              2,
              ourPubkey,
              directPeerPubkeys,
              knownNodeIds,
              knownChannelIds,
              currentNodes,
              currentLinks
            );
            currentNodes = result.nodes;
            currentLinks = result.links;
            hop2Frontier = result.newPeerIds;
            console.info(
              `[NetworkGraph] Hop 2: ${result.newPeerIds.size} new peers, ${currentLinks.length} total links`
            );
          }
        }
      } catch (err) {
        console.error("[NetworkGraph] Failed to fetch gossip for hop 1:", err);
      }

      // For hop 1 peers still without aliases, try mempool/amboss
      const unresolvedPeers = hop1PeerIds.filter((id) => !aliasMap.has(id));
      const aliasPromises = unresolvedPeers.map(async (peerId) => {
        const alias = await fetchNodeAlias(peerId);
        return { peerId, alias };
      });
      const results = await Promise.allSettled(aliasPromises);
      for (const result of results) {
        if (result.status === "fulfilled" && result.value.alias) {
          aliasMap.set(result.value.peerId, result.value.alias);
        }
      }
    }

    // Hop 3+: Gossip expansion from hop 2 frontier outward
    let frontier = [...hop2Frontier];

    for (let hop = 3; hop <= MAX_HOPS; hop++) {
      if (abortRef.current || frontier.length === 0) {
        break;
      }
      if (currentNodes.length >= MAX_NODES) {
        break;
      }

      const BATCH_SIZE = 25;
      const newPeerIds = new Set<string>();

      for (let i = 0; i < frontier.length; i += BATCH_SIZE) {
        if (abortRef.current || currentNodes.length >= MAX_NODES) {
          break;
        }

        const batch = frontier.slice(i, i + BATCH_SIZE);

        try {
          const response = await request<NetworkGraphApiResponse>(
            `/api/node/network-graph?nodeIds=${batch.join(",")}`
          );

          if (!response) {
            continue;
          }

          console.info(
            `[NetworkGraph] Gossip hop ${hop}: ${response.nodes.length} nodes, ${response.channels.length} channels`
          );

          for (const graphNode of response.nodes) {
            const nodeAlias = graphNode.node?.alias || graphNode.node?.Alias;
            if (nodeAlias && !aliasMap.has(graphNode.nodeId)) {
              aliasMap.set(graphNode.nodeId, nodeAlias);
            }
          }

          const result = processGossipChannels(
            response.channels,
            hop,
            ourPubkey,
            directPeerPubkeys,
            knownNodeIds,
            knownChannelIds,
            currentNodes,
            currentLinks
          );
          currentNodes = result.nodes;
          currentLinks = result.links;
          for (const id of result.newPeerIds) {
            newPeerIds.add(id);
          }
        } catch (error) {
          console.error(
            `[NetworkGraph] Failed to fetch hop ${hop} batch:`,
            error
          );
        }
      }

      frontier = [...newPeerIds];
    }

    // Final pass: resolve aliases for nodes that still show truncated pubkeys.
    // First try gossip, then fall back to mempool/amboss for the closest hops.
    const unresolvedNodes = currentNodes
      .filter(
        (n) =>
          !n.isOurNode && !aliasMap.has(n.id) && n.alias === n.id.slice(0, 8)
      )
      .sort((a, b) => a.hop - b.hop);

    if (unresolvedNodes.length > 0 && !abortRef.current) {
      // 1) Batch gossip lookup
      const ALIAS_BATCH = 25;
      const gossipIds = unresolvedNodes.slice(0, 200).map((n) => n.id);

      for (let i = 0; i < gossipIds.length; i += ALIAS_BATCH) {
        if (abortRef.current) {
          break;
        }
        const batch = gossipIds.slice(i, i + ALIAS_BATCH);
        try {
          const response = await request<NetworkGraphApiResponse>(
            `/api/node/network-graph?nodeIds=${batch.join(",")}`
          );
          if (response) {
            for (const graphNode of response.nodes) {
              const nodeAlias = graphNode.node?.alias || graphNode.node?.Alias;
              if (nodeAlias) {
                aliasMap.set(graphNode.nodeId, nodeAlias);
              }
            }
          }
        } catch {
          // ignore
        }
      }

      // 2) For still-unresolved hop 2-3 nodes, try mempool/amboss fallback
      const stillUnresolved = unresolvedNodes
        .filter((n) => !aliasMap.has(n.id) && n.hop <= 3)
        .slice(0, 30);

      if (stillUnresolved.length > 0 && !abortRef.current) {
        const aliasPromises = stillUnresolved.map(async (node) => {
          const alias = await fetchNodeAlias(node.id);
          return { id: node.id, alias };
        });
        const results = await Promise.allSettled(aliasPromises);
        for (const result of results) {
          if (result.status === "fulfilled" && result.value.alias) {
            aliasMap.set(result.value.id, result.value.alias);
          }
        }
      }
    }

    // Apply all collected aliases in one pass
    if (aliasMap.size > 0) {
      currentNodes = currentNodes.map((n) => {
        const alias = aliasMap.get(n.id);
        return alias ? { ...n, alias } : n;
      });
    }

    // Set all state once at the end
    setNodes(currentNodes);
    setLinks(currentLinks);
    setLoading(false);
  }, [ourPubkey]);

  useEffect(() => {
    if (!ourPubkey || !channels) {
      return;
    }

    expand();
    return () => {
      abortRef.current = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- channels triggers initial load only
  }, [expand, !!channels]);

  return { nodes, links, loading };
}
