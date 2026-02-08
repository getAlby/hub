import {
  ArrowDownUpIcon,
  ArrowRightIcon,
  CopyIcon,
  HeartIcon,
  LinkIcon,
  ListIcon,
  Settings2Icon,
  SparklesIcon,
  ZapIcon,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { Link, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import ResponsiveButton from "src/components/ResponsiveButton";
import CircleProgress from "src/components/ui/custom/circle-progress";
import { LinkButton } from "src/components/ui/custom/link-button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { Card } from "src/components/ui/card";
import Loading from "src/components/Loading";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { ONCHAIN_DUST_SATS } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { getNodeHealth } from "src/lib/nodeHealth";
import NetworkGraph from "./NetworkGraph";
import NodeDetailPanel from "./NodeDetailPanel";
import { GraphNode } from "./types";
import { useNetworkGraph } from "./useNetworkGraph";

export default function NetworkGraphPage() {
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: channels } = useChannels();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances(true);
  const { data: albyMe } = useAlbyMe();
  const navigate = useNavigate();
  const { nodes, links, loading } = useNetworkGraph(
    nodeConnectionInfo?.pubkey,
    channels
  );

  const nodeHealth = channels ? getNodeHealth(channels) : 0;

  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const selectedNode = selectedNodeId
    ? (nodes.find((n) => n.id === selectedNodeId) ?? null)
    : null;
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState<{
    width: number;
    height: number;
  } | null>(null);

  // Measure synchronously before first paint so the graph never renders with wrong size
  useLayoutEffect(() => {
    const el = containerRef.current;
    if (el && !dimensions) {
      const rect = el.getBoundingClientRect();
      if (rect.width > 0 && rect.height > 0) {
        setDimensions({ width: rect.width, height: rect.height });
      }
    }
  }, [dimensions]);

  // Track subsequent resizes (window resize, sidebar toggle, etc.)
  useEffect(() => {
    const el = containerRef.current;
    if (!el) {
      return;
    }

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        if (width > 0 && height > 0) {
          setDimensions({ width, height });
        }
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const handleNodeClick = useCallback((node: GraphNode) => {
    setSelectedNodeId((prev) => (prev === node.id ? null : node.id));
  }, []);

  const dataReady = !!nodeConnectionInfo && !!channels;

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <AppHeader
        title="Network Graph"
        contentRight={
          dataReady &&
          hasChannelManagement && (
            <div className="flex gap-3 items-center justify-center">
              <DropdownMenu modal={false}>
                <ResponsiveButton
                  asChild
                  icon={Settings2Icon}
                  text="Advanced"
                  variant="secondary"
                >
                  <DropdownMenuTrigger />
                </ResponsiveButton>
                <DropdownMenuContent className="w-56">
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>Node</DropdownMenuLabel>
                    <DropdownMenuItem>
                      <div
                        className="flex flex-row gap-2 items-center w-full cursor-pointer"
                        onClick={() => {
                          if (!nodeConnectionInfo) {
                            return;
                          }
                          copyToClipboard(nodeConnectionInfo.pubkey);
                        }}
                      >
                        <div>Public key</div>
                        <div className="overflow-hidden text-ellipsis flex-1 text-muted-foreground text-xs">
                          {nodeConnectionInfo?.pubkey || "Loading..."}
                        </div>
                        {nodeConnectionInfo && (
                          <CopyIcon className="shrink-0 size-4" />
                        )}
                      </div>
                    </DropdownMenuItem>
                    {nodeConnectionInfo?.address &&
                      nodeConnectionInfo?.port && (
                        <DropdownMenuItem>
                          <div
                            className="flex flex-row gap-2 items-center w-full cursor-pointer"
                            onClick={() => {
                              const connectionAddress = `${nodeConnectionInfo.pubkey}@${nodeConnectionInfo.address}:${nodeConnectionInfo.port}`;
                              copyToClipboard(connectionAddress);
                            }}
                          >
                            <div>URI</div>
                            <div className="overflow-hidden text-ellipsis flex-1 text-muted-foreground text-xs">
                              {nodeConnectionInfo.pubkey.substring(0, 6)}...@
                              {nodeConnectionInfo.address}:
                              {nodeConnectionInfo.port}
                            </div>
                            <CopyIcon className="shrink-0 size-4" />
                          </div>
                        </DropdownMenuItem>
                      )}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>On-Chain</DropdownMenuLabel>
                    <DropdownMenuItem asChild>
                      <Link
                        to="/channels/onchain/buy-bitcoin"
                        className="w-full"
                      >
                        Buy
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link to="/channels/onchain/deposit-bitcoin">
                        Deposit
                      </Link>
                    </DropdownMenuItem>
                    {(balances?.onchain.spendable || 0) > ONCHAIN_DUST_SATS && (
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/withdraw")}
                      >
                        Withdraw
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  {hasChannelManagement && (
                    <DropdownMenuGroup>
                      <DropdownMenuLabel>Swaps</DropdownMenuLabel>
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/swap?type=in")}
                        className="cursor-pointer"
                      >
                        <div className="mr-2 text-muted-foreground flex flex-row items-center">
                          <LinkIcon className="size-4" />
                          <ArrowRightIcon className="size-4" />
                          <ZapIcon className="size-4" />
                        </div>
                        Swap in
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => navigate("/wallet/swap?type=out")}
                        className="cursor-pointer"
                      >
                        <div className="mr-2 text-muted-foreground flex flex-row items-center">
                          <ZapIcon className="size-4" />
                          <ArrowRightIcon className="size-4" />
                          <LinkIcon className="size-4" />
                        </div>
                        Swap out
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>Management</DropdownMenuLabel>
                    <DropdownMenuItem asChild>
                      <Link className="w-full" to="/peers">
                        Connected Peers
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link className="w-full" to="/wallet/sign-message">
                        Sign Message
                      </Link>
                    </DropdownMenuItem>
                    {info?.backendType === "LDK" &&
                      (!albyMe?.subscription.plan_code ? (
                        <UpgradeDialog>
                          <div className="cursor-pointer">
                            <DropdownMenuItem className="w-full pointer-events-none">
                              <Link
                                className="w-full flex items-center"
                                to="/wallet/node-alias"
                              >
                                <SparklesIcon className="size-4 mr-2" /> Set
                                Node Alias
                              </Link>
                            </DropdownMenuItem>
                          </div>
                        </UpgradeDialog>
                      ) : (
                        <DropdownMenuItem className="w-full">
                          <Link className="w-full" to="/wallet/node-alias">
                            Set Node Alias
                          </Link>
                        </DropdownMenuItem>
                      ))}
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>

              <LinkButton
                to="/channels"
                variant="secondary"
                className="hidden sm:flex"
              >
                <ListIcon />
                Node
              </LinkButton>
              <LinkButton
                to="/wallet/swap"
                variant="secondary"
                className="hidden sm:flex"
              >
                <ArrowDownUpIcon />
                Swap
              </LinkButton>
              <LinkButton to="/channels/incoming">Open Channel</LinkButton>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <ExternalLink to="https://guides.getalby.com/user-guide/alby-hub/node/node-health">
                      <CircleProgress
                        value={nodeHealth}
                        className="w-9 h-9 relative"
                      >
                        {nodeHealth === 100 && (
                          <div className="absolute w-full h-full opacity-20">
                            <div className="absolute w-full h-full bg-primary animate-pulse" />
                          </div>
                        )}
                        <HeartIcon
                          className="size-4"
                          stroke={"var(--color-primary)"}
                          strokeWidth={3}
                          fill={
                            nodeHealth === 100
                              ? "var(--color-primary)"
                              : "transparent"
                          }
                        />
                      </CircleProgress>
                    </ExternalLink>
                  </TooltipTrigger>
                  <TooltipContent>Node health: {nodeHealth}%</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          )
        }
      />
      <Card className="relative flex-1 min-h-0 w-full mt-4 p-0 overflow-hidden">
        <div ref={containerRef} className="absolute inset-0">
          {(!dataReady || loading || nodes.length === 0) && (
            <div className="flex flex-col items-center justify-center h-full gap-3 text-muted-foreground">
              <Loading className="h-8 w-8" />
              <span className="text-sm">Loading network graph...</span>
            </div>
          )}
          {dimensions && dataReady && !loading && nodes.length > 0 && (
            <NetworkGraph
              nodes={nodes}
              links={links}
              onNodeClick={handleNodeClick}
              onDeselect={() => setSelectedNodeId(null)}
              selectedNodeId={selectedNodeId}
              width={dimensions.width}
              height={dimensions.height}
            />
          )}
          {selectedNode && (
            <NodeDetailPanel
              node={selectedNode}
              graphLinks={links}
              graphNodes={nodes}
              onClose={() => setSelectedNodeId(null)}
              onNodeSelect={setSelectedNodeId}
            />
          )}
        </div>
      </Card>
    </div>
  );
}
