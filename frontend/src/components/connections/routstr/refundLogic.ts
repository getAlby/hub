/**
 * Pure decision helpers for the refund drain and auto top-up UI.
 *
 * Extracted from useRoutstrd.ts / RefundDialog.tsx so the money math is
 * unit-testable without React or network access.
 */

/**
 * Amount to melt on a drain pass: the full remaining balance minus the
 * mint's melt fee. Anything the fee leaves behind is melted on a later
 * pass, so the wallet drains to zero.
 */
export function computeRefundSend(balanceSat: number, feeSat: number): number {
  return Math.max(0, Math.floor(balanceSat - feeSat));
}

/**
 * Whether the mint melt fee must be re-quoted: never quoted yet, or the
 * remaining balance moved by at least 10 sats. The fee is flat for small
 * amounts, so near-identical consecutive passes reuse the last quote and
 * avoid creating a throwaway app-scoped invoice per pass.
 */
export function shouldRequoteFee(
  lastQuotedBalance: number,
  balanceSat: number
): boolean {
  return (
    lastQuotedBalance === 0 || Math.abs(balanceSat - lastQuotedBalance) >= 10
  );
}

/**
 * Melt failures that are safe to retry with a smaller send: the known
 * coco-cashu-core degenerate case ("amount must be a non-negative number"
 * when selected proofs exactly equal invoice + fee) and insufficient-proof
 * errors (input fees at small denominations).
 */
export function isRetryableMeltError(message: string): boolean {
  return /insufficient|not enough (funds|proofs)|non-negative/i.test(message);
}

/**
 * Smallest refundable balance: the fee plus 1 sat, never less than 1.
 * A balance at or below the fee cannot be melted (the mint requires
 * proofs >= invoice + fee).
 */
export function computeMinRequired(feeSat: number): number {
  return Math.max(1, feeSat + 1);
}
