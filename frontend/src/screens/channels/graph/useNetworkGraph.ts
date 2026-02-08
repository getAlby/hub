import { useCallback, useEffect, useRef, useState } from "react";
import { Channel, MempoolNode } from "src/types";
import { request } from "src/utils/request";
import {
  GraphLink,
  GraphNode,
  NetworkGraphApiResponse,
  NetworkGraphChannel,
} from "./types";

async function fetchNodeAlias(pubkey: string): Promise<string | null> {
  // Try mempool.space first (same as useNodeDetails hook)
  try {
    const mempool = await request<MempoolNode>(
      `/api/mempool?endpoint=/v1/lightning/nodes/${pubkey}`
    );
    if (mempool?.alias) {
      return mempool.alias;
    }
  } catch {
    // fall through
  }
  // Fallback: try amboss.space GraphQL API
  try {
    const resp = await fetch("https://api.amboss.space/graphql", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: `{ getNode(pubkey: "${pubkey}") { graph_info { node { alias } } } }`,
      }),
    });
    const json = await resp.json();
    const alias = json?.data?.getNode?.graph_info?.node?.alias;
    if (alias) {
      return alias;
    }
  } catch {
    // fall through
  }
  return null;
}

const MAX_HOPS = 6;
const MAX_NODES = 500;
const HOP_COLORS = [
  "hsl(45, 100%, 50%)", // hop 0: our node (gold)
  "hsl(200, 90%, 55%)", // hop 1: direct peers (blue)
  "hsl(160, 70%, 50%)", // hop 2: (teal)
  "hsl(280, 60%, 55%)", // hop 3: (purple)
  "hsl(20, 70%, 55%)", // hop 4: (orange)
  "hsl(340, 60%, 50%)", // hop 5: (pink)
  "hsl(0, 0%, 50%)", // hop 6+: (gray)
];

function getHopColor(hop: number): string {
  return HOP_COLORS[Math.min(hop, HOP_COLORS.length - 1)];
}

function getChannelEndpoints(
  channel: NetworkGraphChannel
): [string, string] | null {
  // Handle LDK (camelCase), LND snake_case, and LND PascalCase
  const node1 = channel.nodeOne || channel.node1_pub || channel.NodeOne;
  const node2 = channel.nodeTwo || channel.node2_pub || channel.NodeTwo;
  if (node1 && node2) {
    return [node1, node2];
  }
  return null;
}

function getChannelCapacity(channel: NetworkGraphChannel): number {
  if (channel.capacitySats != null) {
    return channel.capacitySats;
  }
  if (channel.Capacity != null) {
    return channel.Capacity;
  }
  if (channel.capacity != null) {
    return parseInt(channel.capacity, 10) || 0;
  }
  return 0;
}

function getChannelId(channel: NetworkGraphChannel): string {
  if (channel.shortChannelId != null) {
    return String(channel.shortChannelId);
  }
  if (channel.ChannelId != null) {
    return String(channel.ChannelId);
  }
  if (channel.channel_id) {
    return channel.channel_id;
  }
  const endpoints = getChannelEndpoints(channel);
  if (endpoints) {
    return `${endpoints[0]}-${endpoints[1]}`;
  }
  return Math.random().toString(36);
}

/** Process channels from a gossip API response, adding new nodes and links. */
function processGossipChannels(
  channels: NetworkGraphChannel[],
  hop: number,
  ourPubkey: string,
  directPeerPubkeys: Set<string>,
  knownNodeIds: Set<string>,
  knownChannelIds: Set<string>,
  currentNodes: GraphNode[],
  currentLinks: GraphLink[]
): {
  nodes: GraphNode[];
  links: GraphLink[];
  newPeerIds: Set<string>;
} {
  const newPeerIds = new Set<string>();

  for (const channel of channels) {
    const channelId = getChannelId(channel);
    if (knownChannelIds.has(channelId)) {
      continue;
    }
    knownChannelIds.add(channelId);

    const endpoints = getChannelEndpoints(channel);
    if (!endpoints) {
      continue;
    }

    const [node1, node2] = endpoints;

    // Ensure both endpoint nodes exist in our graph
    for (const nodeId of [node1, node2]) {
      if (!knownNodeIds.has(nodeId)) {
        knownNodeIds.add(nodeId);
        currentNodes = [
          ...currentNodes,
          {
            id: nodeId,
            alias: nodeId.slice(0, 8),
            hop,
            isOurNode: false,
            hasChannel: directPeerPubkeys.has(nodeId),
            color: getHopColor(hop),
          },
        ];
        newPeerIds.add(nodeId);
      }
    }

    currentLinks = [
      ...currentLinks,
      {
        source: node1,
        target: node2,
        capacity: getChannelCapacity(channel),
        isOurChannel: node1 === ourPubkey || node2 === ourPubkey,
      },
    ];
  }

  return { nodes: currentNodes, links: currentLinks, newPeerIds };
}

export function useNetworkGraph(
  ourPubkey: string | undefined,
  channels: Channel[] | undefined
) {
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [links, setLinks] = useState<GraphLink[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentHop, setCurrentHop] = useState(0);
  const [maxHop, setMaxHop] = useState(0);
  const abortRef = useRef(false);

  const expand = useCallback(async () => {
    if (!ourPubkey || !channels) {
      return;
    }

    abortRef.current = false;
    setLoading(true);

    const directPeerPubkeys = new Set(channels.map((c) => c.remotePubkey));

    const knownNodeIds = new Set<string>();
    const knownChannelIds = new Set<string>();
    let currentNodes: GraphNode[] = [];
    let currentLinks: GraphLink[] = [];

    // Hop 0: Our node
    const ourNode: GraphNode = {
      id: ourPubkey,
      alias: "You",
      hop: 0,
      isOurNode: true,
      hasChannel: true,
      color: getHopColor(0),
    };
    knownNodeIds.add(ourPubkey);
    currentNodes = [ourNode];
    setNodes([...currentNodes]);
    setCurrentHop(0);

    // Hop 1: Add direct channel partners from local channel data.
    // This is essential because private channels don't appear in gossip.
    setCurrentHop(1);
    const hop1PeerIds: string[] = [];
    for (const channel of channels) {
      const peerId = channel.remotePubkey;
      if (!knownNodeIds.has(peerId)) {
        knownNodeIds.add(peerId);
        hop1PeerIds.push(peerId);
        currentNodes = [
          ...currentNodes,
          {
            id: peerId,
            alias: peerId.slice(0, 8),
            hop: 1,
            isOurNode: false,
            hasChannel: true,
            color: getHopColor(1),
          },
        ];
      }

      // Add a link for this channel
      const channelId = channel.id;
      if (!knownChannelIds.has(channelId)) {
        knownChannelIds.add(channelId);
        currentLinks = [
          ...currentLinks,
          {
            source: ourPubkey,
            target: peerId,
            capacity: Math.round(
              (channel.localBalance + channel.remoteBalance) / 1000
            ),
            isOurChannel: true,
          },
        ];
      }
    }

    // Fetch gossip for hop 1 peers: resolve aliases AND extract hop 2 channels
    let hop2Frontier = new Set<string>();
    if (hop1PeerIds.length > 0) {
      const unresolvedPeers = new Set(hop1PeerIds);
      try {
        const response = await request<NetworkGraphApiResponse>(
          `/api/node/network-graph?nodeIds=${hop1PeerIds.join(",")}`
        );
        if (response) {
          console.info(
            `[NetworkGraph] Gossip for hop 1 peers: ${response.nodes.length} nodes, ${response.channels.length} channels`
          );

          // Resolve aliases from gossip
          for (const graphNode of response.nodes) {
            const nodeAlias = graphNode.node?.alias || graphNode.node?.Alias;
            if (nodeAlias) {
              currentNodes = currentNodes.map((n) =>
                n.id === graphNode.nodeId ? { ...n, alias: nodeAlias } : n
              );
              unresolvedPeers.delete(graphNode.nodeId);
            }
          }

          // Process channels from gossip to discover hop 2 nodes
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

      // For peers still without aliases, try mempool then amboss
      const aliasPromises = [...unresolvedPeers].map(async (peerId) => {
        const alias = await fetchNodeAlias(peerId);
        return { peerId, alias };
      });
      const results = await Promise.allSettled(aliasPromises);
      for (const result of results) {
        if (result.status === "fulfilled" && result.value.alias) {
          currentNodes = currentNodes.map((n) =>
            n.id === result.value.peerId
              ? { ...n, alias: result.value.alias! }
              : n
          );
        }
      }
    }

    setNodes([...currentNodes]);
    setLinks([...currentLinks]);
    setMaxHop(hop2Frontier.size > 0 ? 2 : 1);

    // Hop 3+: Progressive gossip expansion from hop 2 frontier outward
    // (hop 2 was already processed above from the hop 1 gossip response)
    let frontier = [...hop2Frontier];

    for (let hop = 3; hop <= MAX_HOPS; hop++) {
      if (abortRef.current || frontier.length === 0) {
        break;
      }
      if (currentNodes.length >= MAX_NODES) {
        break;
      }

      setCurrentHop(hop);

      const BATCH_SIZE = 25;
      const newPeerIds = new Set<string>();

      for (let i = 0; i < frontier.length; i += BATCH_SIZE) {
        if (abortRef.current || currentNodes.length >= MAX_NODES) {
          break;
        }

        const batch = frontier.slice(i, i + BATCH_SIZE);
        const nodeIdsParam = batch.join(",");

        try {
          const response = await request<NetworkGraphApiResponse>(
            `/api/node/network-graph?nodeIds=${nodeIdsParam}`
          );

          if (!response) {
            continue;
          }

          console.info(
            `[NetworkGraph] Gossip hop ${hop}: ${response.nodes.length} nodes, ${response.channels.length} channels`
          );

          // Update aliases for newly discovered nodes
          for (const graphNode of response.nodes) {
            const nodeAlias = graphNode.node?.alias || graphNode.node?.Alias;
            if (nodeAlias && knownNodeIds.has(graphNode.nodeId)) {
              currentNodes = currentNodes.map((n) =>
                n.id === graphNode.nodeId && n.alias === n.id.slice(0, 8)
                  ? { ...n, alias: nodeAlias }
                  : n
              );
            }
          }

          // Process channels
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

      // Update state after each hop so the graph visually expands
      setNodes([...currentNodes]);
      setLinks([...currentLinks]);
      setMaxHop(hop);

      // Next frontier: newly discovered nodes
      frontier = [...newPeerIds];

      if (frontier.length === 0) {
        break;
      }
    }

    // Final pass: resolve aliases for nodes that still show truncated pubkeys.
    // First try gossip, then fall back to mempool/amboss for the closest hops.
    const unresolvedNodes = currentNodes
      .filter((n) => !n.isOurNode && n.alias === n.id.slice(0, 8))
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
                currentNodes = currentNodes.map((n) =>
                  n.id === graphNode.nodeId ? { ...n, alias: nodeAlias } : n
                );
              }
            }
          }
        } catch {
          // ignore
        }
      }

      setNodes([...currentNodes]);

      // 2) For still-unresolved hop 2-3 nodes, try mempool/amboss fallback
      const stillUnresolved = currentNodes
        .filter(
          (n) => !n.isOurNode && n.alias === n.id.slice(0, 8) && n.hop <= 3
        )
        .slice(0, 30);

      if (stillUnresolved.length > 0 && !abortRef.current) {
        const aliasPromises = stillUnresolved.map(async (node) => {
          const alias = await fetchNodeAlias(node.id);
          return { id: node.id, alias };
        });
        const results = await Promise.allSettled(aliasPromises);
        for (const result of results) {
          if (result.status === "fulfilled" && result.value.alias) {
            currentNodes = currentNodes.map((n) =>
              n.id === result.value.id
                ? { ...n, alias: result.value.alias! }
                : n
            );
          }
        }
        setNodes([...currentNodes]);
      }
    }

    setLoading(false);
  }, [ourPubkey, channels]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- async data loading pattern
    expand();
    return () => {
      abortRef.current = true;
    };
  }, [expand]);

  return { nodes, links, loading, currentHop, maxHop };
}
