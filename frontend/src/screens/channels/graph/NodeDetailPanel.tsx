import { CopyIcon, ExternalLinkIcon, XIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import { useChannels } from "src/hooks/useChannels";
import { copyToClipboard } from "src/lib/clipboard";
import { GraphNode } from "./types";

type Props = {
  node: GraphNode;
  onClose: () => void;
};

export default function NodeDetailPanel({ node, onClose }: Props) {
  const { data: channels } = useChannels();

  const peerChannels = channels?.filter((c) => c.remotePubkey === node.id);

  return (
    <div className="absolute top-0 right-0 z-20 w-80 max-w-full h-full bg-background border-l border-border p-4 overflow-y-auto">
      <div className="flex justify-between items-start mb-4">
        <h3 className="text-lg font-semibold truncate pr-2">{node.alias}</h3>
        <Button
          variant="ghost"
          size="icon"
          onClick={onClose}
          className="shrink-0"
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

        {/* Channel info for direct peers */}
        {peerChannels && peerChannels.length > 0 && (
          <div>
            <div className="text-xs text-muted-foreground mb-2">Channels</div>
            <div className="space-y-2">
              {peerChannels.map((channel) => {
                const total = channel.localBalance + channel.remoteBalance;
                const localPct =
                  total > 0 ? (channel.localBalance / total) * 100 : 50;
                return (
                  <div
                    key={channel.id}
                    className="border border-border rounded-md p-3 text-xs space-y-2"
                  >
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
      </div>
    </div>
  );
}
