import { describe, expect, it } from "vitest";
import {
  computeMinRequired,
  computeRefundSend,
  isRetryableMeltError,
  shouldRequoteFee,
} from "./refundLogic";

describe("computeRefundSend", () => {
  it("melts balance minus fee", () => {
    expect(computeRefundSend(103, 2)).toBe(101);
    expect(computeRefundSend(50, 2)).toBe(48);
  });

  it("returns 0 when the fee covers the balance", () => {
    expect(computeRefundSend(2, 2)).toBe(0);
    expect(computeRefundSend(1, 2)).toBe(0);
  });

  it("floors fractional results", () => {
    expect(computeRefundSend(50.9, 2)).toBe(48);
  });
});

describe("shouldRequoteFee", () => {
  it("always quotes on the first pass", () => {
    expect(shouldRequoteFee(0, 50)).toBe(true);
  });

  it("reuses the quote when the balance barely moved", () => {
    expect(shouldRequoteFee(100, 105)).toBe(false);
    expect(shouldRequoteFee(100, 109)).toBe(false);
  });

  it("re-quotes when the balance moved 10+ sats", () => {
    expect(shouldRequoteFee(100, 110)).toBe(true);
    expect(shouldRequoteFee(100, 90)).toBe(true);
  });
});

describe("isRetryableMeltError", () => {
  it("matches the degenerate non-negative error", () => {
    expect(
      isRetryableMeltError(
        "500 Payment failed: amount must be a non-negative number"
      )
    ).toBe(true);
  });

  it("matches insufficient proofs/funds", () => {
    expect(
      isRetryableMeltError("Payment failed: Not enough proofs to send")
    ).toBe(true);
    expect(isRetryableMeltError("500 Payment failed: not enough funds")).toBe(
      true
    );
  });

  it("does not retry unrelated failures", () => {
    expect(isRetryableMeltError("mint unreachable")).toBe(false);
    expect(
      isRetryableMeltError("Melt failed: no confirmation from daemon")
    ).toBe(false);
  });
});

describe("computeMinRequired", () => {
  it("is fee + 1", () => {
    expect(computeMinRequired(2)).toBe(3);
    expect(computeMinRequired(0)).toBe(1);
  });
});
