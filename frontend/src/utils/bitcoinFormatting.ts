import { BitcoinDisplayFormat } from "src/types";

/**
 * Utility function to format Bitcoin amounts as a string
 * @param amount - Amount in millisatoshis
 * @param displayFormat - Display format
 * @param showSymbol - Whether to show the symbol/unit
 */
export function formatBitcoinAmount(
  amount: number,
  displayFormat: BitcoinDisplayFormat = "bip177",
  showSymbol: boolean = true
): string {
  const sats = Math.floor(amount / 1000);
  const formattedNumber = new Intl.NumberFormat().format(sats);

  if (!showSymbol) {
    return formattedNumber;
  }

  if (displayFormat === "bip177") {
    return `₿${formattedNumber}`;
  } else {
    return `${formattedNumber} sats`;
  }
}
