import { clsx, type ClassValue } from "clsx";
import { BackendType } from "src/types";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatAmount(amount: number, decimals = 1) {
  amount /= 1000; //msat to sat
  let i = 0;
  for (i; amount >= 1000; i++) {
    amount /= 1000;
  }
  return amount.toFixed(i > 0 ? decimals : 0) + ["", "k", "M", "G"][i];
}

export function splitSocketAddress(socketAddress: string) {
  const lastColonIndex = socketAddress.lastIndexOf(":");
  const address = socketAddress.slice(0, lastColonIndex);
  const port = socketAddress.slice(lastColonIndex + 1);
  return { address, port };
}

export function backendTypeHasMnemonic(backendType: BackendType) {
  return ["LND", "PHOENIX"].indexOf(backendType) === -1;
}
