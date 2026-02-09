import { CopyIcon, ExternalLinkIcon, XIcon } from "lucide-react";
import { useEffect, useMemo } from "react";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import { useChannels } from "src/hooks/useChannels";
import { copyToClipboard } from "src/lib/clipboard";
import { GraphLink, GraphNode } from "./types";

const MAX_VISIBLE_CHANNELS = 50;

type Props = {
  node: GraphNode;
  graphLinks: GraphLink[];
  graphNodes: GraphNode[];
  onClose: () => void;
  onNodeSelect: (nodeId: string) => void;
};

export default function NodeDetailPanel({
  node,
  graphLinks,
  graphNodes,
  onClose,
  onNodeSelect,
}: Props) {
  const { data: channels } = useChannels();

  // Our direct channels: all channels when viewing our node, or filtered by peer
  const peerChannels = node.isOurNode
    ? channels
    : channels?.filter((c) => c.remotePubkey === node.id);

  // Build a node alias lookup for graph channel display
  const nodeAliasMap = useMemo(() => {
    const map = new Map<string, string>();
    for (const n of graphNodes) {
      map.set(n.id, n.alias);
    }
    return map;
  }, [graphNodes]);

  // Dismiss panel on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  // All graph channels involving this node (from gossip data)
  const nodeGraphChannels = useMemo(() => {
    return graphLinks
      .filter((l) => l.source === node.id || l.target === node.id)
      .map((l) => {
        const peerId = l.source === node.id ? l.target : l.source;
        return {
          peerId,
          peerAlias: nodeAliasMap.get(peerId) || peerId.slice(0, 8),
          capacity: l.capacity,
          isOurChannel: l.isOurChannel,
        };
      })
      .sort((a, b) => b.capacity - a.capacity);
  }, [graphLinks, node.id, nodeAliasMap]);

  return (
    <div className="absolute top-0 right-0 z-20 w-80 max-w-full h-full bg-background border-l border-border p-4 overflow-y-auto">
      <div className="flex justify-between items-start mb-4">
        <h3 className="text-lg font-semibold truncate pr-2">{node.alias}</h3>
        <Button
          variant="ghost"
          size="icon"
          onClick={onClose}
          className="shrink-0"
          aria-label="Close panel"
        >
          <XIcon className="size-4" />
        </Button>
      </div>

      <div className="space-y-4">
        {/* Pubkey */}
        <div>
          <div className="text-xs text-muted-foreground mb-1">Public Key</div>
          <div
            className="flex items-center gap-2 cursor-pointer group"
            onClick={() => copyToClipboard(node.id)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                copyToClipboard(node.id);
              }
            }}
            role="button"
            tabIndex={0}
          >
            <span className="text-xs font-mono break-all">{node.id}</span>
            <CopyIcon className="size-3 shrink-0 text-muted-foreground group-hover:text-foreground" />
          </div>
        </div>

        {/* Hop distance */}
        <div>
          <div className="text-xs text-muted-foreground mb-1">
            Network Distance
          </div>
          <div className="flex items-center gap-2">
            <div
              className="size-3 rounded-full"
              style={{ backgroundColor: node.color }}
            />
            <span className="text-sm">
              {node.isOurNode
                ? "Your Node"
                : node.hop === 1
                  ? "Direct peer (1 hop)"
                  : `${node.hop} hops away`}
            </span>
          </div>
        </div>

        {/* Our channels with this peer (detailed balance info) */}
        {peerChannels && peerChannels.length > 0 && (
          <div>
            <div className="text-xs text-muted-foreground mb-2">
              Your Channels
            </div>
            <div className="space-y-2">
              {peerChannels.map((channel) => {
                const total = channel.localBalance + channel.remoteBalance;
                const localPct =
                  total > 0 ? (channel.localBalance / total) * 100 : 50;
                return (
                  <div
                    key={channel.id}
                    className={`border border-border rounded-md p-3 text-xs space-y-2${node.isOurNode ? " cursor-pointer hover:bg-muted/50 transition-colors" : ""}`}
                    onClick={
                      node.isOurNode
                        ? () => onNodeSelect(channel.remotePubkey)
                        : undefined
                    }
                    role={node.isOurNode ? "button" : undefined}
                    tabIndex={node.isOurNode ? 0 : undefined}
                    onKeyDown={
                      node.isOurNode
                        ? (e) => {
                            if (e.key === "Enter" || e.key === " ") {
                              e.preventDefault();
                              onNodeSelect(channel.remotePubkey);
                            }
                          }
                        : undefined
                    }
                  >
                    {node.isOurNode && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Peer</span>
                        <span className="truncate ml-2">
                          {nodeAliasMap.get(channel.remotePubkey) ||
                            channel.remotePubkey.slice(0, 8)}
                        </span>
                      </div>
                    )}
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Status</span>
                      <span
                        className={
                          channel.status === "online"
                            ? "text-green-500"
                            : channel.status === "opening"
                              ? "text-yellow-500"
                              : "text-red-500"
                        }
                      >
                        {channel.status}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Capacity</span>
                      <span>
                        {Math.round(total / 1000).toLocaleString()} sats
                      </span>
                    </div>
                    <div>
                      <div className="flex justify-between text-muted-foreground mb-1">
                        <span>Local</span>
                        <span>Remote</span>
                      </div>
                      <div className="w-full h-2 rounded-full bg-muted overflow-hidden">
                        <div
                          className="h-full bg-primary rounded-full"
                          style={{ width: `${localPct}%` }}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* External links */}
        {!node.isOurNode && (
          <div className="pt-2 space-y-2">
            <ExternalLink
              to={`https://amboss.space/node/${node.id}`}
              className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
            >
              View on Amboss
              <ExternalLinkIcon className="size-3" />
            </ExternalLink>
          </div>
        )}

        {/* All graph channels for this node */}
        {nodeGraphChannels.length > 0 && (
          <div>
            <div className="text-xs text-muted-foreground mb-2">
              Known Channels ({nodeGraphChannels.length})
            </div>
            <div className="space-y-1.5">
              {nodeGraphChannels.slice(0, MAX_VISIBLE_CHANNELS).map((ch, i) => (
                <div
                  key={`${ch.peerId}-${i}`}
                  className="flex items-center justify-between text-xs border border-border rounded-md px-3 py-2 cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => onNodeSelect(ch.peerId)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      onNodeSelect(ch.peerId);
                    }
                  }}
                >
                  <span className="truncate mr-2">
                    {ch.isOurChannel && !node.isOurNode ? (
                      <span className="font-semibold text-primary">You</span>
                    ) : (
                      ch.peerAlias
                    )}
                  </span>
                  <span className="text-muted-foreground whitespace-nowrap">
                    {ch.capacity.toLocaleString()} sats
                  </span>
                </div>
              ))}
              {nodeGraphChannels.length > MAX_VISIBLE_CHANNELS && (
                <div className="text-xs text-muted-foreground text-center py-1">
                  {nodeGraphChannels.length - MAX_VISIBLE_CHANNELS} more
                  channels not shown
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
