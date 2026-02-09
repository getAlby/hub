import { act, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { GraphLink, GraphNode } from "./types";

// ─── ForceGraph2D ref method mocks ──────────────────────────────────
const mockCenterAt = vi.fn();
const mockZoom = vi.fn();
const mockZoomToFit = vi.fn();
const mockGetGraphBbox = vi.fn();
const mockD3Force = vi.fn((name: string) => {
  if (name === "link") {
    return {
      distance: vi.fn().mockReturnThis(),
      strength: vi.fn().mockReturnThis(),
      iterations: vi.fn().mockReturnThis(),
    };
  }
  if (name === "charge") {
    return {
      strength: vi.fn().mockReturnThis(),
      distanceMax: vi.fn().mockReturnThis(),
    };
  }
  return undefined;
});
const mockScreen2GraphCoords = vi.fn(() => ({ x: 0, y: 0 }));

// Ref-like container to capture ForceGraph2D props without variable reassignment
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const captured: { props: Record<string, any> } = { props: {} };

vi.mock("react-force-graph-2d", async () => {
  const React = await import("react");
  const ForceGraph2D = React.forwardRef(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (props: any, ref: any) => {
      React.useImperativeHandle(ref, () => ({
        centerAt: mockCenterAt,
        zoom: mockZoom,
        zoomToFit: mockZoomToFit,
        getGraphBbox: mockGetGraphBbox,
        d3Force: mockD3Force,
        screen2GraphCoords: mockScreen2GraphCoords,
      }));
      React.useEffect(() => {
        captured.props = props;
      });
      return <div data-testid="force-graph-2d" />;
    }
  );
  ForceGraph2D.displayName = "ForceGraph2D";
  return { default: ForceGraph2D };
});

vi.mock("src/components/ui/theme-provider", () => ({
  useTheme: () => ({ isDarkMode: false }),
}));

vi.mock("src/components/ui/button", () => ({
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  Button: ({ children, onClick, title }: any) => (
    <button data-testid="center-button" onClick={onClick} title={title}>
      {children}
    </button>
  ),
}));

vi.mock("lucide-react", () => ({
  LocateFixedIcon: () => <span />,
}));

vi.mock("./graphUtils", () => ({
  HOP_COLORS: [
    "hsl(45, 100%, 50%)",
    "hsl(200, 90%, 55%)",
    "hsl(160, 70%, 50%)",
  ],
  HOP_LABELS: ["Your node", "Direct peers", "2 hops"],
}));

import NetworkGraph from "./NetworkGraph";

// ─── Fixtures ───────────────────────────────────────────────────────

const OUR_NODE: GraphNode = {
  id: "our-node",
  alias: "My Node",
  hop: 0,
  isOurNode: true,
  hasChannel: true,
  color: "hsl(45, 100%, 50%)",
};

const PEER_NODE: GraphNode = {
  id: "peer-1",
  alias: "Peer One",
  hop: 1,
  isOurNode: false,
  hasChannel: true,
  color: "hsl(200, 80%, 55%)",
};

const REMOTE_NODE: GraphNode = {
  id: "remote-1",
  alias: "Remote Node",
  hop: 2,
  isOurNode: false,
  hasChannel: false,
  color: "hsl(280, 60%, 55%)",
};

const LINK_OUR_PEER: GraphLink = {
  source: "our-node",
  target: "peer-1",
  capacity: 1000000,
  isOurChannel: true,
};

const LINK_PEER_REMOTE: GraphLink = {
  source: "peer-1",
  target: "remote-1",
  capacity: 500000,
  isOurChannel: false,
};

// Must match DETAIL_PANEL_WIDTH in NetworkGraph.tsx
const DETAIL_PANEL_WIDTH = 320;

function makeProps(
  overrides: Partial<Parameters<typeof NetworkGraph>[0]> = {}
) {
  return {
    nodes: [OUR_NODE, PEER_NODE, REMOTE_NODE],
    links: [LINK_OUR_PEER, LINK_PEER_REMOTE],
    onNodeClick: vi.fn(),
    onDeselect: vi.fn(),
    selectedNodeId: null as string | null,
    width: 1000,
    height: 600,
    ...overrides,
  };
}

describe("NetworkGraph", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
    vi.restoreAllMocks();
    captured.props = {};
  });

  /** Render and advance past the initial-center polling interval (50 x 100ms). */
  function renderAndSettle(props: ReturnType<typeof makeProps>) {
    const result = render(<NetworkGraph {...props} />);
    act(() => {
      vi.advanceTimersByTime(5100);
    });
    // Clear mock calls from initial render / settling
    mockCenterAt.mockClear();
    mockZoom.mockClear();
    mockZoomToFit.mockClear();
    mockGetGraphBbox.mockClear();
    return result;
  }

  it("renders ForceGraph2D with correct dimensions", () => {
    renderAndSettle(makeProps());
    expect(screen.getByTestId("force-graph-2d")).toBeInTheDocument();
    expect(captured.props.width).toBe(1000);
    expect(captured.props.height).toBe(600);
  });

  it("displays node and channel counts", () => {
    renderAndSettle(makeProps());
    expect(screen.getByText(/3 nodes.*2 channels/)).toBeInTheDocument();
  });

  // ─── applyZoomToFit via node click ────────────────────────────────

  describe("applyZoomToFit via node click", () => {
    it("computes correct centerAt and zoom accounting for panel offset", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue({ x: [0, 100], y: [0, 50] });

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...OUR_NODE, x: 50, y: 25 });
      });

      // pad=80, bboxW=100, bboxH=50
      // visibleWidth = 1000 - 320 = 680
      // zoomX = (680-160)/100 = 5.2, zoomY = (600-160)/50 = 8.8
      // targetZoom = clamp(min(5.2, 8.8), 0.1, 8) = 5.2
      const expectedZoom = 5.2;
      // cx=50, cy=25, panelOffset = 320/(2*5.2)
      const expectedCx = 50 + DETAIL_PANEL_WIDTH / 2 / expectedZoom;

      expect(mockZoom).toHaveBeenCalledWith(expectedZoom);
      const [actualCx, actualCy] = mockCenterAt.mock.calls[0];
      expect(actualCx).toBeCloseTo(expectedCx, 5);
      expect(actualCy).toBe(25);
    });

    it("calls onNodeClick with the matching GraphNode", () => {
      const onNodeClick = vi.fn();
      const props = makeProps({ onNodeClick });
      mockGetGraphBbox.mockReturnValue({ x: [0, 100], y: [0, 50] });

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...PEER_NODE, x: 100, y: 50 });
      });

      expect(onNodeClick).toHaveBeenCalledWith(PEER_NODE);
    });

    it("filters bbox to selected node and direct neighbors only", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue({ x: [0, 200], y: [0, 100] });

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...OUR_NODE, x: 0, y: 0 });
      });

      const lastCall =
        mockGetGraphBbox.mock.calls[mockGetGraphBbox.mock.calls.length - 1];
      const filterFn = lastCall[0];

      // our-node (selected) -> included
      expect(filterFn({ id: "our-node" })).toBe(true);
      // peer-1 (neighbor via LINK_OUR_PEER) -> included
      expect(filterFn({ id: "peer-1" })).toBe(true);
      // remote-1 is NOT a direct neighbor of our-node -> excluded
      expect(filterFn({ id: "remote-1" })).toBe(false);
    });

    it("clamps zoom to minimum 0.1 for very large bounding boxes", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue({
        x: [-50000, 50000],
        y: [-50000, 50000],
      });

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...OUR_NODE, x: 0, y: 0 });
      });

      expect(mockZoom).toHaveBeenCalledWith(0.1);
    });

    it("clamps zoom to maximum 8 for very small bounding boxes", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue({
        x: [49.5, 50.5],
        y: [24.5, 25.5],
      });

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...OUR_NODE, x: 50, y: 25 });
      });

      expect(mockZoom).toHaveBeenCalledWith(8);
    });

    it("does not call centerAt or zoom when getGraphBbox returns null", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue(null);

      renderAndSettle(props);

      act(() => {
        captured.props.onNodeClick({ ...OUR_NODE, x: 50, y: 25 });
      });

      expect(mockCenterAt).not.toHaveBeenCalled();
      expect(mockZoom).not.toHaveBeenCalled();
    });
  });

  // ─── applyZoomToFit via selectedNodeId prop change ────────────────

  describe("applyZoomToFit via selectedNodeId prop", () => {
    it("fires zoom via rAF when selectedNodeId changes", () => {
      const props = makeProps();
      mockGetGraphBbox.mockReturnValue({ x: [0, 100], y: [0, 50] });

      const { rerender } = renderAndSettle(props);

      rerender(<NetworkGraph {...props} selectedNodeId="our-node" />);

      // Advance past the requestAnimationFrame
      act(() => {
        vi.advanceTimersByTime(16);
      });

      expect(mockGetGraphBbox).toHaveBeenCalled();
      expect(mockCenterAt).toHaveBeenCalled();
      expect(mockZoom).toHaveBeenCalledWith(5.2);
    });
  });

  // ─── Center button ────────────────────────────────────────────────

  describe("center button", () => {
    it("calls zoomToFit(500, 60) when clicked", () => {
      renderAndSettle(makeProps());
      mockZoomToFit.mockClear();

      const btn = screen.getByTitle("Center on your node");
      act(() => {
        btn.click();
      });

      expect(mockZoomToFit).toHaveBeenCalledWith(500, 60);
    });
  });

  // ─── Background click ─────────────────────────────────────────────

  describe("background click", () => {
    it("calls onDeselect when no node is near the click", () => {
      const onDeselect = vi.fn();
      const props = makeProps({ onDeselect });
      // Return coords far from any node
      mockScreen2GraphCoords.mockReturnValue({ x: 9999, y: 9999 });

      renderAndSettle(props);

      act(() => {
        captured.props.onBackgroundClick({ offsetX: 500, offsetY: 300 });
      });

      expect(onDeselect).toHaveBeenCalled();
    });
  });
});
