import { useInfo } from "src/hooks/useInfo";
import { BitcoinDisplayFormat } from "src/utils/bitcoinFormatting";

interface FormattedBitcoinAmountProps {
  amount: number; // Amount in millisatoshis
  className?: string;
  showSymbol?: boolean; // Whether to show the symbol/unit
}

/**
 * Formats a Bitcoin amount according to user settings
 * @param amount - Amount in millisatoshis
 * @param className - Optional CSS classes
 * @param showSymbol - Whether to show the symbol/unit (default: true)
 */
export function FormattedBitcoinAmount({
  amount,
  className = "",
  showSymbol = true,
}: FormattedBitcoinAmountProps) {
  const { data: info } = useInfo();

  // Convert from millisatoshis to satoshis
  const sats = Math.floor(amount / 1000);

  // Get display format from settings, default to BIP177
  const displayFormat: BitcoinDisplayFormat =
    info?.bitcoinDisplayFormat || "bip177";

  const formattedNumber = new Intl.NumberFormat().format(sats);

  if (!showSymbol) {
    return <span className={className}>{formattedNumber}</span>;
  }

  if (displayFormat === "bip177") {
    return <span className={className}>₿{formattedNumber}</span>;
  } else {
    return <span className={className}>{formattedNumber} sats</span>;
  }
}
