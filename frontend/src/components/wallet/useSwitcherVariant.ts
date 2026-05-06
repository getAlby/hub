import { useLocation } from "react-router";

const STORAGE_KEY = "balance-switcher-variant";
export type Variant = "icons" | "tabs" | "original";

export function useSwitcherVariant(): [Variant, (v: Variant) => void] {
  const stored =
    (typeof window !== "undefined" &&
      (localStorage.getItem(STORAGE_KEY) as Variant)) ||
    "icons";
  const set = (v: Variant) => {
    localStorage.setItem(STORAGE_KEY, v);
    window.location.reload();
  };
  return [stored, set];
}

export function useActiveSwitcher() {
  const { pathname } = useLocation();
  return pathname.startsWith("/wallet/onchain") ? "onchain" : "lightning";
}
