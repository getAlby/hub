import { MempoolNode } from "src/types";
import { request } from "src/utils/request";
import { GraphLink, GraphNode, NetworkGraphChannel } from "./types";

export const MAX_HOPS = 6;
export const MAX_NODES = 500;

export const HOP_COLORS = [
  "hsl(45, 100%, 50%)", // hop 0: our node (gold)
  "hsl(200, 90%, 55%)", // hop 1: direct peers (blue)
  "hsl(160, 70%, 50%)", // hop 2: (teal)
  "hsl(280, 60%, 55%)", // hop 3: (purple)
  "hsl(20, 70%, 55%)", // hop 4: (orange)
  "hsl(340, 60%, 50%)", // hop 5: (pink)
  "hsl(0, 0%, 50%)", // hop 6+: (gray)
];

export const HOP_LABELS = [
  "Your node",
  "Direct peers",
  "2 hops",
  "3 hops",
  "4 hops",
  "5 hops",
  "6+ hops",
];

export function getHopColor(hop: number): string {
  return HOP_COLORS[Math.min(hop, HOP_COLORS.length - 1)];
}

export function getChannelEndpoints(
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

export function getChannelCapacity(channel: NetworkGraphChannel): number {
  // LDK PascalCase: CapacitySats is *uint64 (nullable)
  if (channel.CapacitySats != null) {
    return channel.CapacitySats;
  }
  // LDK camelCase (older versions)
  if (channel.capacitySats != null) {
    return channel.capacitySats;
  }
  // LND PascalCase
  if (channel.Capacity != null) {
    return channel.Capacity;
  }
  // LND snake_case
  if (channel.capacity != null) {
    return parseInt(channel.capacity, 10) || 0;
  }
  // Fallback: estimate from max HTLC in routing policies (msat → sats)
  const maxHtlc = Math.max(
    channel.OneToTwo?.HtlcMaximumMsat ?? 0,
    channel.TwoToOne?.HtlcMaximumMsat ?? 0
  );
  if (maxHtlc > 0) {
    return Math.round(maxHtlc / 1000);
  }
  return 0;
}

export function getChannelId(channel: NetworkGraphChannel): string {
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
  // Deterministic fallback from available fields to avoid breaking deduplication
  const stable = JSON.stringify(channel);
  let hash = 0;
  for (let i = 0; i < stable.length; i++) {
    hash = (hash * 31 + stable.charCodeAt(i)) | 0;
  }
  return `unknown-${hash.toString(36)}`;
}

export async function fetchNodeAlias(pubkey: string): Promise<string | null> {
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
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);
    const resp = await fetch("https://api.amboss.space/graphql", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: `query GetNode($pubkey: String!) { getNode(pubkey: $pubkey) { graph_info { node { alias } } } }`,
        variables: { pubkey },
      }),
      signal: controller.signal,
    });
    clearTimeout(timeoutId);
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

/** Process channels from a gossip API response, adding new nodes and links. */
export function processGossipChannels(
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
  const newNodes: GraphNode[] = [];
  const newLinks: GraphLink[] = [];

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
        newNodes.push({
          id: nodeId,
          alias: nodeId.slice(0, 8),
          hop,
          isOurNode: false,
          hasChannel: directPeerPubkeys.has(nodeId),
          color: getHopColor(hop),
        });
        newPeerIds.add(nodeId);
      }
    }

    newLinks.push({
      source: node1,
      target: node2,
      capacity: getChannelCapacity(channel),
      isOurChannel: node1 === ourPubkey || node2 === ourPubkey,
    });
  }

  return {
    nodes: [...currentNodes, ...newNodes],
    links: [...currentLinks, ...newLinks],
    newPeerIds,
  };
}
