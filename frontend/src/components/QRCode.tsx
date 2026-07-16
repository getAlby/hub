import { type ReactNode, useEffect, useMemo, useRef } from "react";
import QRCodeStyling, { type Options } from "qr-code-styling";
import { BitcoinPaymentIcon } from "src/components/icons/BitcoinPayment";
import { LightningIcon } from "src/components/icons/Lightning";
import { cn } from "src/lib/utils";

export type Props = {
  value: string;
  size?: number;
  className?: string;
  showAvatar?: boolean;
  frameType?: "lightning" | "onchain";
  paymentType?: "lightning" | "onchain";
  centerContent?: ReactNode;

  // Use Q when an external overlay covers part of the QR code.
  level?: "Q" | undefined;
};

function QRCode({
  value,
  size = 256,
  level,
  className,
  showAvatar = false,
  frameType,
  paymentType,
  centerContent,
}: Props) {
  const resolvedFrameType = paymentType ?? frameType;
  const hasCenterContent = Boolean(centerContent);
  const containerRef = useRef<HTMLDivElement>(null);
  const options = useMemo<Options>(
    () => ({
      type: "svg",
      width: size,
      height: size,
      data: value,
      image: showAvatar ? "/icon-lightmode.svg" : undefined,
      margin: 0,
      qrOptions: {
        errorCorrectionLevel:
          level ?? (showAvatar || paymentType || hasCenterContent ? "Q" : "M"),
      },
      imageOptions: {
        crossOrigin: "anonymous",
        hideBackgroundDots: true,
        imageSize: 0.15,
        margin: 4,
      },
      dotsOptions: {
        color: "var(--qr-foreground)",
        type: "dots",
        roundSize: false,
      },
      cornersSquareOptions: {
        color: "var(--qr-foreground)",
        type: "extra-rounded",
      },
      cornersDotOptions: {
        color: "var(--qr-foreground)",
        type: "dot",
      },
      backgroundOptions: {
        color: "var(--qr-background)",
      },
    }),
    [hasCenterContent, level, paymentType, showAvatar, size, value]
  );

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const qrCode = new QRCodeStyling(options);
    qrCode.append(container);

    return () => {
      container.replaceChildren();
    };
  }, [options]);

  return (
    <div
      className={cn(
        "relative w-full rounded-[28px] p-2",
        resolvedFrameType === "lightning"
          ? "bg-payment-lightning"
          : resolvedFrameType === "onchain"
            ? "bg-payment-onchain"
            : "bg-primary",
        className
      )}
      style={{ maxWidth: size }}
    >
      <div className="rounded-3xl bg-qr-background p-4">
        <div
          ref={containerRef}
          className="aspect-square w-full overflow-hidden [&_svg]:size-full"
        />
      </div>
      {(paymentType || centerContent) && (
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-qr-background p-1 text-qr-background">
          {centerContent ??
            (paymentType === "lightning" ? (
              <LightningIcon className="size-12" />
            ) : (
              <BitcoinPaymentIcon className="size-12" />
            ))}
        </div>
      )}
    </div>
  );
}

export default QRCode;
