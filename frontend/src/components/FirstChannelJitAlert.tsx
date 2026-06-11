import { AlertTriangleIcon, InfoIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { CreateInvoiceRequest, Transaction } from "src/types";
import { request } from "src/utils/request";

const PROBE_TIMEOUT_MS = 5000;

export default function FirstChannelJitAlert() {
  const { data: info } = useInfo();
  const { data: channels } = useChannels();
  const { data: balances } = useBalances();

  // a JIT channel only opens when the feature is enabled AND an LSPS2 liquidity
  // source is actually configured (jitChannelsEnabled alone is just a settings
  // toggle and can be true on backends without an LSPS2 source).
  const lsps2Source = info?.jitChannelsEnabled
    ? info.jitChannelsLiquiditySource
    : undefined;

  const minPaymentSizeMsat = info?.jitChannelsMinPaymentSizeMsat;

  const isJitEnabled = !!lsps2Source && !!channels;
  // the user's first received payment opens the channel when they have none yet.
  const isFirstChannel = isJitEnabled && channels.length === 0;

  // probe whether a JIT channel can actually be obtained by requesting an
  // invoice for the minimum payment size. If it (or waiting for the minimum
  // payment size) doesn't succeed within the timeout, we surface a fallback
  // alert depending on whether the user already has channels.
  const [probeState, setProbeState] = React.useState<
    "loading" | "ok" | "failed"
  >("loading");
  const deadlineRef = React.useRef<number | null>(null);
  // single-flight the non-idempotent probe invoice: hold the in-flight request
  // so effect re-entry (e.g. StrictMode remount) reuses the same POST instead
  // of creating a duplicate invoice. Reset when the probe window ends.
  const probeRequestRef = React.useRef<Promise<Transaction | undefined> | null>(
    null
  );

  React.useEffect(() => {
    if (!isJitEnabled) {
      deadlineRef.current = null;
      probeRequestRef.current = null;
      return;
    }

    // start the 5s clock once when we enter JIT mode - it keeps ticking while
    // we wait for the minimum payment size to become available.
    if (deadlineRef.current === null) {
      deadlineRef.current = Date.now() + PROBE_TIMEOUT_MS;
    }

    let cancelled = false;
    const remainingMs = deadlineRef.current - Date.now();
    if (remainingMs <= 0) {
      setProbeState("failed");
      return;
    }

    const timer = setTimeout(() => {
      if (!cancelled) {
        setProbeState("failed");
      }
    }, remainingMs);

    // wait for the minimum payment size before requesting the probe invoice.
    if (minPaymentSizeMsat) {
      // reuse an already in-flight probe so a re-run doesn't issue a second POST.
      const probeRequest =
        probeRequestRef.current ??
        (probeRequestRef.current = request<Transaction>("/api/invoices", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            amountMsat: minPaymentSizeMsat,
            description: "",
          } as CreateInvoiceRequest),
        }));
      probeRequest
        .then(() => {
          if (!cancelled) {
            clearTimeout(timer);
            setProbeState("ok");
          }
        })
        .catch(() => {
          if (!cancelled) {
            clearTimeout(timer);
            setProbeState("failed");
          }
        });
    }

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [isJitEnabled, minPaymentSizeMsat]);

  if (!isJitEnabled || probeState === "loading") {
    return null;
  }

  if (probeState === "failed") {
    // no channels yet and a JIT channel couldn't be obtained - the user has no
    // receiving capacity at all and will fail to receive until a channel opens.
    if (isFirstChannel) {
      return (
        <Alert variant="warning">
          <AlertTriangleIcon className="h-4 w-4" />
          <AlertTitle>Can't receive payments yet</AlertTitle>
          <AlertDescription className="inline">
            You won't be able to receive payments until you{" "}
            <Link className="underline" to="/channels/incoming">
              open a channel
            </Link>
            .
          </AlertDescription>
        </Alert>
      );
    }

    // they already have channels but a JIT channel couldn't be obtained, so they
    // can only receive up to their current capacity without opening one.
    return (
      <Alert>
        <InfoIcon className="h-4 w-4" />
        <AlertDescription className="inline">
          You can currently receive up to{" "}
          <FormattedBitcoinAmount
            amountMsat={balances?.lightning.totalReceivableMsat ?? 0}
          />
          . If you want to receive a larger payment,{" "}
          <Link className="underline" to="/channels/incoming">
            open a channel
          </Link>
          .
        </AlertDescription>
      </Alert>
    );
  }

  // probe succeeded - only the first-channel case needs an informational alert.
  if (!isFirstChannel) {
    return null;
  }

  return (
    <Alert>
      <InfoIcon className="h-4 w-4" />
      <AlertTitle>First payment opens a channel</AlertTitle>
      <AlertDescription className="inline">
        A channel fee applies.{" "}
        {!!minPaymentSizeMsat && (
          <>
            Minimum payment{" "}
            <FormattedBitcoinAmount amountMsat={minPaymentSizeMsat} />.{" "}
          </>
        )}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/faq/what-are-just-in-time-channels"
          className="underline"
        >
          Learn more
        </ExternalLink>
      </AlertDescription>
    </Alert>
  );
}
