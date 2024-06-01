import ReactQRCode from "react-qr-code";
import { useTheme } from "src/components/ui/theme-provider";
import { cn } from "src/lib/utils";


export type Props = {
  value: string;
  size?: number;
  className?: string;

  // set the level to Q if there are overlays
  // Q will improve error correction (so we can add overlays covering up to 25% of the QR)
  // at the price of decreased information density (meaning the QR codes "pixels" have to be
  // smaller to encode the same information).
  // While that isn't that much of a problem for lightning addresses (because they are usually quite short),
  // for invoices that contain larger amount of data those QR codes can get "harder" to read.
  // (meaning you have to aim your phone very precisely and have to wait longer for the reader
  // to recognize the QR code)
  level?: "Q" | undefined;
};

function QRCode({ value, size, level, className }: Props) {
  const { theme } = useTheme();
  const fgColor = theme === "dark" ? "#FFFFFF" : "#242424";
  const bgColor = theme === "dark" ? "#242424" : "#FFFFFF";

  return (
    <ReactQRCode
      value={value}
      size={size}
      fgColor={fgColor}
      bgColor={bgColor}
      className={cn("rounded-md ph-no-capture", className)}
      level={level}
    />
  );
}

export default QRCode;
