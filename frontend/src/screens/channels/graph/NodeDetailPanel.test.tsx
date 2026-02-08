import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import NodeDetailPanel from "./NodeDetailPanel";
import { GraphLink, GraphNode } from "./types";

// Mock useChannels hook
vi.mock("src/hooks/useChannels", () => ({
  useChannels: vi.fn(() => ({ data: undefined })),
}));

// Mock clipboard utility
vi.mock("src/lib/clipboard", () => ({
  copyToClipboard: vi.fn(),
}));

// Mock isHttpMode so ExternalLink renders as <Link>
vi.mock("src/utils/isHttpMode", () => ({
  isHttpMode: () => true,
}));

function renderPanel({
  node,
  graphLinks = [],
  graphNodes = [],
}: {
  node: GraphNode;
  graphLinks?: GraphLink[];
  graphNodes?: GraphNode[];
}) {
  return render(
    <MemoryRouter>
      <NodeDetailPanel
        node={node}
        graphLinks={graphLinks}
        graphNodes={graphNodes}
        onClose={() => {}}
        onNodeSelect={() => {}}
      />
    </MemoryRouter>
  );
}

const ourNode: GraphNode = {
  id: "our-pubkey-abcdef1234567890",
  alias: "MyNode",
  hop: 0,
  isOurNode: true,
  hasChannel: true,
  color: "hsl(45, 100%, 50%)",
};

const remoteNode: GraphNode = {
  id: "remote-pubkey-abcdef1234567890",
  alias: "RemotePeer",
  hop: 3,
  isOurNode: false,
  hasChannel: false,
  color: "hsl(280, 60%, 55%)",
};

const directPeerNode: GraphNode = {
  id: "peer-pubkey-abcdef1234567890",
  alias: "DirectPeer",
  hop: 1,
  isOurNode: false,
  hasChannel: true,
  color: "hsl(200, 90%, 55%)",
};

describe("NodeDetailPanel", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders node alias and pubkey", () => {
    renderPanel({ node: remoteNode });

    expect(screen.getByText("RemotePeer")).toBeInTheDocument();
    expect(
      screen.getByText("remote-pubkey-abcdef1234567890")
    ).toBeInTheDocument();
  });

  it("shows 'Your Node' for our node", () => {
    renderPanel({ node: ourNode });

    expect(screen.getByText("Your Node")).toBeInTheDocument();
  });

  it("shows hop distance for remote nodes", () => {
    renderPanel({ node: remoteNode });

    expect(screen.getByText("3 hops away")).toBeInTheDocument();
  });

  it("shows 'You' for our channels when viewing a non-our node", () => {
    const links: GraphLink[] = [
      {
        source: ourNode.id,
        target: directPeerNode.id,
        capacity: 100000,
        isOurChannel: true,
      },
    ];
    const nodes: GraphNode[] = [ourNode, directPeerNode];

    renderPanel({
      node: directPeerNode,
      graphLinks: links,
      graphNodes: nodes,
    });

    expect(screen.getByText("You")).toBeInTheDocument();
  });

  it("shows peer alias for our node's known channels", () => {
    const links: GraphLink[] = [
      {
        source: ourNode.id,
        target: directPeerNode.id,
        capacity: 100000,
        isOurChannel: true,
      },
    ];
    const nodes: GraphNode[] = [ourNode, directPeerNode];

    renderPanel({
      node: ourNode,
      graphLinks: links,
      graphNodes: nodes,
    });

    // When viewing our own node, the channel peer alias should show
    expect(screen.getByText("DirectPeer")).toBeInTheDocument();
  });

  it("shows capacity values", () => {
    const links: GraphLink[] = [
      {
        source: remoteNode.id,
        target: directPeerNode.id,
        capacity: 500000,
        isOurChannel: false,
      },
    ];
    const nodes: GraphNode[] = [ourNode, remoteNode, directPeerNode];

    renderPanel({
      node: remoteNode,
      graphLinks: links,
      graphNodes: nodes,
    });

    expect(screen.getByText("500,000 sats")).toBeInTheDocument();
  });

  it("shows 'Your Channels' section for direct peers", async () => {
    // Mock useChannels to return a channel for this peer
    const { useChannels } = await import("src/hooks/useChannels");
    vi.mocked(useChannels).mockReturnValue({
      data: [
        {
          id: "ch1",
          remotePubkey: directPeerNode.id,
          localBalance: 300000,
          remoteBalance: 200000,
          status: "online",
          active: true,
          public: true,
          fundingTxId: "tx1",
          confirmations: 6,
          confirmationsRequired: 6,
          forwardingFeeBaseMsat: 1000,
          unspendablePunishmentReserve: 10000,
          counterpartyUnspendablePunishmentReserve: 10000,
        },
      ],
      error: undefined,
      isLoading: false,
      isValidating: false,
      mutate: vi.fn(),
    } as never);

    renderPanel({ node: directPeerNode });

    expect(screen.getByText("Your Channels")).toBeInTheDocument();
  });

  it("shows Amboss link for non-our nodes only", () => {
    // Remote node should have the link
    const { unmount } = renderPanel({ node: remoteNode });
    expect(screen.getByText("View on Amboss")).toBeInTheDocument();
    unmount();

    // Our node should not
    renderPanel({ node: ourNode });
    expect(screen.queryByText("View on Amboss")).not.toBeInTheDocument();
  });
});
