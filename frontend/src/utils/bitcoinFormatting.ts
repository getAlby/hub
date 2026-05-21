import { BITCOIN_DISPLAY_FORMAT_BIP177 } from "src/constants";
import { BitcoinDisplayFormat } from "src/types";

/**
 * Utility function to format Bitcoin amounts as a string
 * @param amountMsat - Amount in millisatoshis
 * @param displayFormat - Display format (required)
 * @param showSymbol - Whether to show the symbol/unit
 */
export function formatBitcoinAmount(
  amountMsat: number,
  displayFormat: BitcoinDisplayFormat,
  showSymbol: boolean = true
): string {
  const sats = Math.floor(amountMsat / 1000);
  const formattedNumber = new Intl.NumberFormat().format(sats);

  if (!showSymbol) {
    return formattedNumber;
  }

  if (displayFormat === BITCOIN_DISPLAY_FORMAT_BIP177) {
    return `₿${formattedNumber}`;
  } else {
    return `${formattedNumber} sats`;
  }
}
