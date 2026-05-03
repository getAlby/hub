export type Bip21Data = {
  address: string;
  amountSat?: number;
  label?: string;
  message?: string;
  lightning?: string; // BOLT11 invoice
};

export function parseBip21(uri: string): Bip21Data {
  // Strip the bitcoin: scheme (handles bitcoin:, bitcoin://, BITCOIN:, etc.)
  const withoutScheme = uri.replace(/^bitcoin:\/?\/?/i, "");
  if (!withoutScheme) {
    throw new Error("Invalid BIP21 URI: missing address");
  }

  const separatorIndex = withoutScheme.indexOf("?");
  const address =
    separatorIndex >= 0
      ? withoutScheme.slice(0, separatorIndex)
      : withoutScheme;

  const result: Bip21Data = { address };

  if (separatorIndex >= 0) {
    const params = new URLSearchParams(withoutScheme.slice(separatorIndex + 1));

    const amountBtc = params.get("amount");
    if (amountBtc) {
      result.amountSat = Math.round(parseFloat(amountBtc) * 100_000_000);
    }

    const label = params.get("label");
    if (label) {
      result.label = label;
    }

    const message = params.get("message");
    if (message) {
      result.message = message;
    }

    const lightning = params.get("lightning");
    if (lightning) {
      result.lightning = lightning;
    }
  }

  return result;
}
