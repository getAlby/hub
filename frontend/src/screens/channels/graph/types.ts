export type GraphNode = {
  id: string;
  alias: string;
  hop: number;
  isOurNode: boolean;
  hasChannel: boolean;
  color: string;
};

export type GraphLink = {
  source: string;
  target: string;
  capacity: number;
  isOurChannel: boolean;
};

// The network-graph API returns nodes with their gossip info and channels.
// The exact shape of `node` varies by backend (LDK vs LND), so we keep it flexible.
export type NetworkGraphNode = {
  nodeId: string;
  node: {
    alias?: string;
    Alias?: string;
    channels?: number[];
    // LND fields
    pub_key?: string;
    PubKey?: string;
  } | null;
};

// Routing policy info from LDK gossip
export type ChannelUpdateInfo = {
  HtlcMaximumMsat?: number;
  HtlcMinimumMsat?: number;
  Fees?: {
    BaseMsat?: number;
    ProportionalMillionths?: number;
  };
};

export type NetworkGraphChannel = {
  // LDK fields (PascalCase from Go struct without json tags)
  NodeOne?: string;
  OneToTwo?: ChannelUpdateInfo | null;
  NodeTwo?: string;
  TwoToOne?: ChannelUpdateInfo | null;
  CapacitySats?: number | null;
  // LDK fields (camelCase — older versions)
  shortChannelId?: number;
  capacitySats?: number;
  nodeOne?: string;
  nodeTwo?: string;
  // LND fields (snake_case from protobuf json tags)
  channel_id?: string;
  chan_point?: string;
  capacity?: string;
  node1_pub?: string;
  node2_pub?: string;
  // LND fields (PascalCase)
  ChannelId?: number;
  Capacity?: number;
};

export type NetworkGraphApiResponse = {
  nodes: NetworkGraphNode[];
  channels: NetworkGraphChannel[];
};
