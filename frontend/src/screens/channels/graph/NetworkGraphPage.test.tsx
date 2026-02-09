import { act, render, screen } from "@testing-library/react";
import {
  afterEach,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";
import { MemoryRouter } from "react-router-dom";
import NetworkGraphPage from "./NetworkGraphPage";

// Polyfill ResizeObserver for jsdom
beforeAll(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver;
});

// Mock getBoundingClientRect before each test so the graph container has dimensions
beforeEach(() => {
  vi.spyOn(Element.prototype, "getBoundingClientRect").mockReturnValue({
    width: 800,
    height: 600,
    top: 0,
    left: 0,
    bottom: 600,
    right: 800,
    x: 0,
    y: 0,
    toJSON: () => {},
  });
});

// Mock AppHeader to avoid SidebarProvider dependency — render title only
vi.mock("src/components/AppHeader", () => ({
  default: ({ title }: { title: string }) => (
    <header data-testid="app-header">
      <h1>{title}</h1>
    </header>
  ),
}));

// Mock Loading component
vi.mock("src/components/Loading", () => ({
  default: () => <div data-testid="loading-spinner" />,
}));

// Mock hooks
vi.mock("src/hooks/useNodeConnectionInfo", () => ({
  useNodeConnectionInfo: vi.fn(() => ({
    data: { pubkey: "our-pubkey-abc123", address: "127.0.0.1", port: 9735 },
  })),
}));

vi.mock("src/hooks/useChannels", () => ({
  useChannels: vi.fn(() => ({
    data: [
      {
        id: "ch1",
        remotePubkey: "peer-pubkey-123",
        localBalance: 500000,
        remoteBalance: 500000,
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
  })),
}));

vi.mock("src/hooks/useInfo", () => ({
  useInfo: vi.fn(() => ({
    data: { backendType: "LDK" },
    hasChannelManagement: true,
  })),
}));

vi.mock("src/hooks/useBalances", () => ({
  useBalances: vi.fn(() => ({
    data: { onchain: { spendable: 0 } },
  })),
}));

vi.mock("src/hooks/useAlbyMe", () => ({
  useAlbyMe: vi.fn(() => ({ data: undefined })),
}));

vi.mock("src/lib/clipboard", () => ({
  copyToClipboard: vi.fn(),
}));

vi.mock("src/utils/isHttpMode", () => ({
  isHttpMode: () => true,
}));

// Mock useNetworkGraph with controllable loading state
const mockUseNetworkGraph = vi.fn();
vi.mock("./useNetworkGraph", () => ({
  useNetworkGraph: (...args: unknown[]) => mockUseNetworkGraph(...args),
}));

// Store the onReady callback so tests can trigger it
let capturedOnReady: (() => void) | undefined;

// Mock NetworkGraph component to avoid canvas rendering in jsdom
vi.mock("./NetworkGraph", () => ({
  default: (props: { onReady?: () => void }) => {
    capturedOnReady = props.onReady;
    return <div data-testid="network-graph">NetworkGraph</div>;
  },
}));

function renderPage() {
  return render(
    <MemoryRouter>
      <NetworkGraphPage />
    </MemoryRouter>
  );
}

describe("NetworkGraphPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    capturedOnReady = undefined;
  });

  it("shows loading overlay while data is loading", () => {
    mockUseNetworkGraph.mockReturnValue({
      nodes: [],
      links: [],
      loading: true,
    });

    renderPage();
    expect(screen.getByText("Loading network graph...")).toBeInTheDocument();
  });

  it("shows loading overlay until onReady fires (loading=false, graph rendered)", () => {
    mockUseNetworkGraph.mockReturnValue({
      nodes: [
        {
          id: "our-pubkey-abc123",
          alias: "You",
          hop: 0,
          isOurNode: true,
          hasChannel: true,
          color: "hsl(45, 100%, 50%)",
        },
      ],
      links: [],
      loading: false,
    });

    renderPage();

    // Loading overlay should still be visible because onReady hasn't been called
    expect(screen.getByText("Loading network graph...")).toBeInTheDocument();
    // The NetworkGraph component should be rendered underneath
    expect(screen.getByTestId("network-graph")).toBeInTheDocument();
  });

  it("hides loading overlay after onReady is called", () => {
    mockUseNetworkGraph.mockReturnValue({
      nodes: [
        {
          id: "our-pubkey-abc123",
          alias: "You",
          hop: 0,
          isOurNode: true,
          hasChannel: true,
          color: "hsl(45, 100%, 50%)",
        },
      ],
      links: [],
      loading: false,
    });

    renderPage();

    // Loading overlay should be visible before onReady
    expect(screen.getByText("Loading network graph...")).toBeInTheDocument();
    expect(capturedOnReady).toBeDefined();

    // Simulate the graph signaling ready
    act(() => {
      capturedOnReady!();
    });

    // After onReady, the loading overlay should be gone
    expect(
      screen.queryByText("Loading network graph...")
    ).not.toBeInTheDocument();
  });

  it("renders 'Network Graph' title", () => {
    mockUseNetworkGraph.mockReturnValue({
      nodes: [],
      links: [],
      loading: false,
    });

    renderPage();
    expect(screen.getByText("Network Graph")).toBeInTheDocument();
  });

  it("does not render NetworkGraph when there are no nodes", () => {
    mockUseNetworkGraph.mockReturnValue({
      nodes: [],
      links: [],
      loading: false,
    });

    renderPage();
    expect(screen.queryByTestId("network-graph")).not.toBeInTheDocument();
  });
});
