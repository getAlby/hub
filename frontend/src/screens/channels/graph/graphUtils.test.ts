import { describe, expect, it } from "vitest";
import { NetworkGraphChannel } from "./types";
import {
  getChannelCapacity,
  getChannelEndpoints,
  getChannelId,
  getHopColor,
  HOP_COLORS,
  processGossipChannels,
} from "./graphUtils";

describe("getChannelCapacity", () => {
  it("returns LDK PascalCase CapacitySats", () => {
    const ch: NetworkGraphChannel = { CapacitySats: 50000 };
    expect(getChannelCapacity(ch)).toBe(50000);
  });

  it("returns 0 for CapacitySats: 0", () => {
    const ch: NetworkGraphChannel = { CapacitySats: 0 };
    expect(getChannelCapacity(ch)).toBe(0);
  });

  it("falls back to HtlcMaximumMsat when CapacitySats is null", () => {
    const ch: NetworkGraphChannel = {
      CapacitySats: null,
      OneToTwo: { HtlcMaximumMsat: 100_000 },
      TwoToOne: { HtlcMaximumMsat: 200_000 },
    };
    expect(getChannelCapacity(ch)).toBe(200);
  });

  it("returns LDK camelCase capacitySats", () => {
    const ch: NetworkGraphChannel = { capacitySats: 75000 };
    expect(getChannelCapacity(ch)).toBe(75000);
  });

  it("returns LND PascalCase Capacity", () => {
    const ch: NetworkGraphChannel = { Capacity: 100000 };
    expect(getChannelCapacity(ch)).toBe(100000);
  });

  it("parses LND snake_case capacity string", () => {
    const ch: NetworkGraphChannel = { capacity: "250000" };
    expect(getChannelCapacity(ch)).toBe(250000);
  });

  it("returns 0 for non-numeric capacity string", () => {
    const ch: NetworkGraphChannel = { capacity: "not-a-number" };
    expect(getChannelCapacity(ch)).toBe(0);
  });

  it("returns 0 when all fields are missing", () => {
    const ch: NetworkGraphChannel = {};
    expect(getChannelCapacity(ch)).toBe(0);
  });
});

describe("getChannelEndpoints", () => {
  it("returns LDK camelCase endpoints", () => {
    const ch: NetworkGraphChannel = { nodeOne: "aaa", nodeTwo: "bbb" };
    expect(getChannelEndpoints(ch)).toEqual(["aaa", "bbb"]);
  });

  it("returns LND snake_case endpoints", () => {
    const ch: NetworkGraphChannel = { node1_pub: "ccc", node2_pub: "ddd" };
    expect(getChannelEndpoints(ch)).toEqual(["ccc", "ddd"]);
  });

  it("returns LDK PascalCase endpoints", () => {
    const ch: NetworkGraphChannel = { NodeOne: "eee", NodeTwo: "fff" };
    expect(getChannelEndpoints(ch)).toEqual(["eee", "fff"]);
  });

  it("returns null when both endpoints are missing", () => {
    const ch: NetworkGraphChannel = {};
    expect(getChannelEndpoints(ch)).toBeNull();
  });

  it("returns null when only one endpoint is provided", () => {
    const ch: NetworkGraphChannel = { nodeOne: "aaa" };
    expect(getChannelEndpoints(ch)).toBeNull();
  });
});

describe("getChannelId", () => {
  it("returns shortChannelId when present", () => {
    const ch: NetworkGraphChannel = {
      shortChannelId: 12345,
      ChannelId: 99999,
    };
    expect(getChannelId(ch)).toBe("12345");
  });

  it("returns ChannelId when shortChannelId is missing", () => {
    const ch: NetworkGraphChannel = { ChannelId: 67890 };
    expect(getChannelId(ch)).toBe("67890");
  });

  it("returns channel_id string", () => {
    const ch: NetworkGraphChannel = { channel_id: "abc123" };
    expect(getChannelId(ch)).toBe("abc123");
  });

  it("falls back to endpoint concatenation", () => {
    const ch: NetworkGraphChannel = { nodeOne: "pubA", nodeTwo: "pubB" };
    expect(getChannelId(ch)).toBe("pubA-pubB");
  });

  it("returns deterministic string when all fields are missing", () => {
    const ch: NetworkGraphChannel = {};
    const id1 = getChannelId(ch);
    const id2 = getChannelId(ch);
    expect(typeof id1).toBe("string");
    expect(id1).toMatch(/^unknown-/);
    expect(id1).toBe(id2); // deterministic
  });

  it("prefers shortChannelId over ChannelId and channel_id", () => {
    const ch: NetworkGraphChannel = {
      shortChannelId: 111,
      ChannelId: 222,
      channel_id: "333",
    };
    expect(getChannelId(ch)).toBe("111");
  });
});

describe("getHopColor", () => {
  it("returns gold for hop 0", () => {
    expect(getHopColor(0)).toBe(HOP_COLORS[0]);
  });

  it("returns purple for hop 3", () => {
    expect(getHopColor(3)).toBe(HOP_COLORS[3]);
  });

  it("clamps to last color for hop beyond range", () => {
    expect(getHopColor(10)).toBe(HOP_COLORS[HOP_COLORS.length - 1]);
  });
});

describe("processGossipChannels", () => {
  const ourPubkey = "our-pubkey";
  const directPeerPubkeys = new Set(["peer1"]);

  function makeContext() {
    return {
      knownNodeIds: new Set<string>([ourPubkey]),
      knownChannelIds: new Set<string>(),
      currentNodes: [
        {
          id: ourPubkey,
          alias: "You",
          hop: 0,
          isOurNode: true,
          hasChannel: true,
          color: getHopColor(0),
        },
      ],
      currentLinks: [] as {
        source: string;
        target: string;
        capacity: number;
        isOurChannel: boolean;
      }[],
    };
  }

  it("adds new nodes and links for a valid channel", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      {
        shortChannelId: 1,
        nodeOne: "peer1",
        nodeTwo: "peer2",
        CapacitySats: 50000,
      },
    ];

    const result = processGossipChannels(
      channels,
      2,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    expect(result.nodes.length).toBe(3); // our node + peer1 + peer2
    expect(result.links.length).toBe(1);
    expect(result.links[0].capacity).toBe(50000);
    expect(result.newPeerIds.size).toBe(2);
  });

  it("deduplicates channels by ID", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      { shortChannelId: 1, nodeOne: "a", nodeTwo: "b", CapacitySats: 100 },
      { shortChannelId: 1, nodeOne: "a", nodeTwo: "b", CapacitySats: 100 },
    ];

    const result = processGossipChannels(
      channels,
      2,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    expect(result.links.length).toBe(1);
  });

  it("deduplicates nodes", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      { shortChannelId: 1, nodeOne: "a", nodeTwo: "b", CapacitySats: 100 },
      { shortChannelId: 2, nodeOne: "a", nodeTwo: "c", CapacitySats: 200 },
    ];

    const result = processGossipChannels(
      channels,
      2,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    // a appears in both channels but should only be added once
    const aNodes = result.nodes.filter((n) => n.id === "a");
    expect(aNodes.length).toBe(1);
  });

  it("marks isOurChannel when our pubkey is an endpoint", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      {
        shortChannelId: 1,
        nodeOne: ourPubkey,
        nodeTwo: "peer1",
        CapacitySats: 100,
      },
    ];

    const result = processGossipChannels(
      channels,
      1,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    expect(result.links[0].isOurChannel).toBe(true);
  });

  it("sets hasChannel for direct peers", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      {
        shortChannelId: 1,
        nodeOne: "peer1",
        nodeTwo: "remote",
        CapacitySats: 100,
      },
    ];

    const result = processGossipChannels(
      channels,
      2,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    const peer1Node = result.nodes.find((n) => n.id === "peer1");
    const remoteNode = result.nodes.find((n) => n.id === "remote");
    expect(peer1Node?.hasChannel).toBe(true);
    expect(remoteNode?.hasChannel).toBe(false);
  });

  it("returns newPeerIds for newly discovered nodes", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      { shortChannelId: 1, nodeOne: "x", nodeTwo: "y", CapacitySats: 100 },
    ];

    const result = processGossipChannels(
      channels,
      3,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    expect(result.newPeerIds.has("x")).toBe(true);
    expect(result.newPeerIds.has("y")).toBe(true);
  });

  it("skips channels with invalid endpoints", () => {
    const ctx = makeContext();
    const channels: NetworkGraphChannel[] = [
      { shortChannelId: 1, CapacitySats: 100 }, // no endpoints
    ];

    const result = processGossipChannels(
      channels,
      2,
      ourPubkey,
      directPeerPubkeys,
      ctx.knownNodeIds,
      ctx.knownChannelIds,
      ctx.currentNodes,
      ctx.currentLinks
    );

    expect(result.links.length).toBe(0);
    expect(result.newPeerIds.size).toBe(0);
  });
});
