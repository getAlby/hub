import { BITCOIN_DISPLAY_FORMAT_BIP177 } from "src/constants";
import { useInfo } from "src/hooks/useInfo";

interface FormattedBitcoinAmountProps {
  amountMsat: number;
  className?: string;
  showSymbol?: boolean; // Whether to show the symbol/unit
}

export function FormattedBitcoinAmount({
  amountMsat,
  className = "",
  showSymbol = true,
}: FormattedBitcoinAmountProps) {
  const { data: info } = useInfo();

  if (!info) {
    return null;
  }

  const sats = Math.floor(amountMsat / 1000);

  // Get display format from settings
  const displayFormat = info.bitcoinDisplayFormat;

  const formattedNumber = new Intl.NumberFormat().format(sats);

  if (!showSymbol) {
    return <span className={className}>{formattedNumber}</span>;
  }

  if (displayFormat === BITCOIN_DISPLAY_FORMAT_BIP177) {
    return <span className={className}>₿{formattedNumber}</span>;
  } else {
    return <span className={className}>{formattedNumber} sats</span>;
  }
}
