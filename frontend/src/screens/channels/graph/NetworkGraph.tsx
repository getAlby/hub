import { LocateFixedIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef } from "react";
import ForceGraph2D, {
  ForceGraphMethods,
  NodeObject,
} from "react-force-graph-2d";
import { Button } from "src/components/ui/button";
import { HOP_COLORS, HOP_LABELS } from "./graphUtils";
import { GraphLink, GraphNode } from "./types";

// After d3-force processes links, source/target become node objects
type ProcessedGraphLink = Omit<GraphLink, "source" | "target"> & {
  source: string | { id: string };
  target: string | { id: string };
};

// d3-force types for configuring simulation forces
interface D3LinkForce {
  distance(d: number | ((link: GraphLink) => number)): D3LinkForce;
  strength(s: number | ((link: GraphLink) => number)): D3LinkForce;
  iterations(n: number): D3LinkForce;
}
interface D3ChargeForce {
  strength(s: number): D3ChargeForce;
  distanceMax(d: number): D3ChargeForce;
}

type Props = {
  nodes: GraphNode[];
  links: GraphLink[];
  loading: boolean;
  currentHop: number;
  maxHop: number;
  onNodeClick: (node: GraphNode) => void;
  onDeselect: () => void;
  selectedNodeId: string | null;
  width: number;
  height: number;
};

type NodeType = NodeObject<GraphNode>;

function getNodeRadius(graphNode: GraphNode): number {
  if (graphNode.isOurNode) {
    return 14;
  }
  if (graphNode.hasChannel) {
    return 10;
  }
  return 5;
}

export default function NetworkGraph({
  nodes,
  links,
  loading,
  currentHop,
  maxHop,
  onNodeClick,
  onDeselect,
  selectedNodeId,
  width,
  height,
}: Props) {
  const graphRef = useRef<ForceGraphMethods<NodeType>>();
  const initialCenterDone = useRef(false);
  const ourNodeRef = useRef<NodeType | null>(null);
  const isDragging = useRef(false);

  // Build a lookup map for fast node access in canvas callbacks
  const nodeMap = useMemo(() => {
    const map = new Map<string | number, GraphNode>();
    for (const n of nodes) {
      map.set(n.id, n);
    }
    return map;
  }, [nodes]);

  // Compute highlighted neighbor set when a node is selected
  const highlightedNodeIds = useMemo(() => {
    if (!selectedNodeId) {
      return new Set<string>();
    }
    const set = new Set<string>();
    set.add(selectedNodeId);
    for (const link of links) {
      if (link.source === selectedNodeId || link.target === selectedNodeId) {
        set.add(link.source);
        set.add(link.target);
      }
    }
    return set;
  }, [selectedNodeId, links]);

  // Configure d3-force simulation for springy drag physics.
  // Re-apply when graphData changes since the simulation may reinitialize.
  useEffect(() => {
    if (!graphRef.current) {
      return;
    }

    const linkForce = graphRef.current.d3Force("link") as
      | D3LinkForce
      | undefined;
    if (linkForce) {
      linkForce
        .distance((link: GraphLink) => (link.isOurChannel ? 150 : 250))
        .strength((link: GraphLink) => (link.isOurChannel ? 0.5 : 0.15))
        .iterations(3);
    }

    const chargeForce = graphRef.current.d3Force("charge") as
      | D3ChargeForce
      | undefined;
    if (chargeForce) {
      chargeForce.strength(-500).distanceMax(1000);
    }
  }, [nodes, links]);

  // Center on our node: poll until ourNodeRef has coordinates after first render
  const prevNodeCount = useRef(0);
  useEffect(() => {
    if (!graphRef.current || nodes.length === 0) {
      return;
    }

    if (!initialCenterDone.current) {
      // Poll until ourNodeRef has position data (set by nodeCanvasObject on first paint)
      let attempts = 0;
      const interval = setInterval(() => {
        attempts++;
        if (ourNodeRef.current?.x != null && graphRef.current) {
          graphRef.current.centerAt(
            ourNodeRef.current.x,
            ourNodeRef.current.y,
            0
          );
          graphRef.current.zoom(1.5, 0);
          initialCenterDone.current = true;
          prevNodeCount.current = nodes.length;
          clearInterval(interval);
        } else if (attempts >= 50) {
          graphRef.current?.centerAt(0, 0, 0);
          graphRef.current?.zoom(1.5, 0);
          initialCenterDone.current = true;
          prevNodeCount.current = nodes.length;
          clearInterval(interval);
        }
      }, 100);
      return () => clearInterval(interval);
    }

    // Re-center when new hop data arrives during loading
    const nodeCountChanged = nodes.length !== prevNodeCount.current;
    prevNodeCount.current = nodes.length;
    if (loading && nodeCountChanged) {
      setTimeout(() => {
        if (ourNodeRef.current?.x != null && graphRef.current) {
          graphRef.current.centerAt(
            ourNodeRef.current.x,
            ourNodeRef.current.y,
            400
          );
        }
      }, 300);
    }
  }, [nodes, loading]);

  // Build graphData early so findNearestNode can access d3-mutated node positions
  const graphData = useMemo(
    () => ({ nodes: [...nodes], links: links.map((l) => ({ ...l })) }),
    [nodes, links]
  );

  const centerOnOurNode = useCallback(() => {
    if (ourNodeRef.current && graphRef.current) {
      graphRef.current.centerAt(
        ourNodeRef.current.x,
        ourNodeRef.current.y,
        500
      );
      graphRef.current.zoom(1.5, 500);
    }
  }, []);

  // Find the nearest node to a graph-space coordinate within a threshold
  const findNearestNode = useCallback(
    (gx: number, gy: number, threshold: number) => {
      let closest: { node: NodeType; dist: number } | null = null;
      for (const node of graphData.nodes as NodeType[]) {
        if (node.x == null || node.y == null) {
          continue;
        }
        const dx = node.x - gx;
        const dy = node.y - gy;
        const dist = Math.sqrt(dx * dx + dy * dy);
        if (dist < threshold && (!closest || dist < closest.dist)) {
          closest = { node, dist };
        }
      }
      return closest?.node ?? null;
    },
    [graphData.nodes]
  );

  const selectAndCenter = useCallback(
    (node: NodeType) => {
      const graphNode = node.id != null ? nodeMap.get(node.id) : undefined;
      if (graphNode) {
        onNodeClick(graphNode);
        if (node.x != null && node.y != null && graphRef.current) {
          graphRef.current.centerAt(node.x, node.y, 500);
        }
      }
    },
    [nodeMap, onNodeClick]
  );

  const handleNodeClick = useCallback(
    (node: NodeType) => {
      selectAndCenter(node);
    },
    [selectAndCenter]
  );

  const nodeCanvasObject = useCallback(
    (node: NodeType, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const graphNode = node.id != null ? nodeMap.get(node.id) : undefined;
      if (!graphNode || node.x == null || node.y == null) {
        return;
      }

      // Track our node's position for centering
      if (graphNode.isOurNode) {
        ourNodeRef.current = node;
      }

      const isSelected = selectedNodeId === graphNode.id;
      const isHighlighted = highlightedNodeIds.has(graphNode.id);
      const isDimmed = selectedNodeId != null && !isHighlighted;
      const radius = getNodeRadius(graphNode);

      // Draw selection ring
      if (isSelected) {
        ctx.beginPath();
        ctx.arc(node.x, node.y, radius + 4, 0, 2 * Math.PI);
        ctx.fillStyle = "rgba(255, 255, 255, 0.3)";
        ctx.fill();
      }

      // Draw highlight ring for neighbors of selected node
      if (isHighlighted && !isSelected && selectedNodeId != null) {
        ctx.beginPath();
        ctx.arc(node.x, node.y, radius + 3, 0, 2 * Math.PI);
        ctx.fillStyle = "rgba(255, 255, 255, 0.15)";
        ctx.fill();
      }

      // Draw node circle
      ctx.beginPath();
      ctx.arc(node.x, node.y, radius, 0, 2 * Math.PI);
      ctx.globalAlpha = isDimmed ? 0.15 : 1;
      ctx.fillStyle = graphNode.color;
      ctx.fill();

      if (graphNode.isOurNode || isSelected) {
        ctx.strokeStyle = "white";
        ctx.lineWidth = 2;
        ctx.stroke();
      }

      // Draw label: always for our node and direct peers, others when zoomed
      const showLabel =
        graphNode.isOurNode ||
        graphNode.hasChannel ||
        isSelected ||
        isHighlighted ||
        globalScale > 1.5;

      if (showLabel && !isDimmed) {
        const fontSize = Math.max(12 / globalScale, 2);
        const label =
          graphNode.alias.length > 20
            ? graphNode.alias.slice(0, 18) + "..."
            : graphNode.alias;

        ctx.font = `${graphNode.isOurNode || graphNode.hasChannel ? "bold " : ""}${fontSize}px sans-serif`;
        ctx.textAlign = "center";
        ctx.textBaseline = "top";

        // Text shadow for readability
        ctx.fillStyle = "rgba(0, 0, 0, 0.7)";
        ctx.fillText(label, node.x + 0.5, node.y + radius + 3 + 0.5);
        ctx.fillStyle = "rgba(255, 255, 255, 0.95)";
        ctx.fillText(label, node.x, node.y + radius + 3);
      }

      ctx.globalAlpha = 1;
    },
    [nodeMap, selectedNodeId, highlightedNodeIds]
  );

  const nodePointerAreaPaint = useCallback(
    (node: NodeType, color: string, ctx: CanvasRenderingContext2D) => {
      const graphNode = node.id != null ? nodeMap.get(node.id) : undefined;
      if (!graphNode || node.x == null || node.y == null) {
        return;
      }
      const radius = getNodeRadius(graphNode);
      // When a node is selected, dimmed nodes get exact-size hit areas
      // to prevent their generous hit areas from intercepting clicks
      // meant for highlighted nodes. Generous default hit areas since
      // onBackgroundClick has a fallback nearest-node finder.
      let hitRadius: number;
      if (selectedNodeId != null) {
        const isHighlighted = highlightedNodeIds.has(graphNode.id);
        hitRadius = isHighlighted ? Math.max(radius + 20, 35) : radius;
      } else {
        hitRadius = Math.max(radius + 20, 35);
      }
      ctx.beginPath();
      ctx.arc(node.x, node.y, hitRadius, 0, 2 * Math.PI);
      ctx.fillStyle = color;
      ctx.fill();
    },
    [nodeMap, selectedNodeId, highlightedNodeIds]
  );

  const linkColor = useCallback(
    (link: ProcessedGraphLink) => {
      if (selectedNodeId) {
        const sourceId =
          typeof link.source === "object" ? link.source.id : link.source;
        const targetId =
          typeof link.target === "object" ? link.target.id : link.target;
        const isConnected =
          sourceId === selectedNodeId || targetId === selectedNodeId;
        if (isConnected) {
          return link.isOurChannel
            ? "rgba(255, 200, 50, 0.9)"
            : "rgba(255, 255, 255, 0.7)";
        }
        return "rgba(255, 255, 255, 0.02)";
      }
      return link.isOurChannel
        ? "rgba(255, 200, 50, 0.6)"
        : "rgba(255, 255, 255, 0.1)";
    },
    [selectedNodeId]
  );

  const linkWidth = useCallback(
    (link: ProcessedGraphLink) => {
      if (selectedNodeId) {
        const sourceId =
          typeof link.source === "object" ? link.source.id : link.source;
        const targetId =
          typeof link.target === "object" ? link.target.id : link.target;
        const isConnected =
          sourceId === selectedNodeId || targetId === selectedNodeId;
        if (isConnected) {
          return link.isOurChannel ? 3 : 2;
        }
        return 0.2;
      }
      if (link.isOurChannel) {
        return 2.5;
      }
      const cap = link.capacity || 1;
      return Math.max(0.3, Math.min(2, Math.log10(cap) / 6));
    },
    [selectedNodeId]
  );

  // Compute legend entries from actual hop depth
  const legendEntries = useMemo(() => {
    const entries: { color: string; label: string }[] = [];
    for (let i = 0; i <= Math.min(maxHop, HOP_COLORS.length - 1); i++) {
      entries.push({ color: HOP_COLORS[i], label: HOP_LABELS[i] });
    }
    return entries;
  }, [maxHop]);

  return (
    <div className="relative w-full h-full">
      <ForceGraph2D
        ref={graphRef}
        graphData={graphData}
        nodeId="id"
        linkSource="source"
        linkTarget="target"
        width={width}
        height={height}
        backgroundColor="transparent"
        nodeCanvasObjectMode={() => "replace"}
        nodeCanvasObject={nodeCanvasObject}
        nodePointerAreaPaint={nodePointerAreaPaint}
        linkColor={linkColor}
        linkWidth={linkWidth}
        onNodeClick={handleNodeClick}
        onBackgroundClick={(event: MouseEvent) => {
          // Fallback: if library's hit detection missed, find nearest node manually
          if (graphRef.current) {
            const { x, y } = graphRef.current.screen2GraphCoords(
              event.offsetX,
              event.offsetY
            );
            const nearest = findNearestNode(x, y, 40);
            if (nearest) {
              selectAndCenter(nearest);
              return;
            }
          }
          onDeselect();
        }}
        nodeLabel=""
        cooldownTicks={100000}
        d3AlphaDecay={0.03}
        d3VelocityDecay={0.4}
        warmupTicks={200}
        onNodeDrag={() => {
          // Reheat once at drag start for immediate spring response
          if (!isDragging.current) {
            isDragging.current = true;
            graphRef.current?.d3ReheatSimulation();
          }
        }}
        onNodeDragEnd={(node: NodeType) => {
          isDragging.current = false;
          // Clear fixed position so the node can bounce back freely
          node.fx = undefined;
          node.fy = undefined;
          // Reheat for bounce-back spring effect
          graphRef.current?.d3ReheatSimulation();
        }}
        enableNodeDrag={true}
        enableZoomInteraction={true}
        enablePanInteraction={true}
        minZoom={0.1}
        maxZoom={10}
      />

      {/* Color legend */}
      <div className="absolute top-4 right-4 z-10 bg-background/80 backdrop-blur-sm rounded-md px-3 py-2 text-xs">
        {legendEntries.map((entry) => (
          <div key={entry.label} className="flex items-center gap-2 py-0.5">
            <span
              className="inline-block w-2.5 h-2.5 rounded-full flex-shrink-0"
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-muted-foreground">{entry.label}</span>
          </div>
        ))}
      </div>

      {/* Center button */}
      <Button
        variant="secondary"
        size="icon"
        className="absolute bottom-4 right-4 z-10"
        onClick={centerOnOurNode}
        title="Center on your node"
      >
        <LocateFixedIcon className="size-4" />
      </Button>

      {/* Hop indicator */}
      {loading && (
        <div className="absolute top-4 left-4 z-10 bg-background/80 backdrop-blur-sm rounded-md px-3 py-1.5 text-sm text-muted-foreground">
          Exploring hop {currentHop}...
        </div>
      )}

      {/* Node count */}
      <div className="absolute bottom-4 left-4 z-10 bg-background/80 backdrop-blur-sm rounded-md px-3 py-1.5 text-xs text-muted-foreground">
        {nodes.length} nodes &middot; {links.length} channels
      </div>
    </div>
  );
}
