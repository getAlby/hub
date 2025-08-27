import { clsx, type ClassValue } from "clsx";
import { BudgetRenewalType } from "src/types";
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

export function generatePageNumbers(currentPage: number, totalPages: number) {
  const MAX_PAGES_TO_SHOW = 3;
  const pageNumbers: (number | "ellipsis")[] = [];
  const half = Math.floor(MAX_PAGES_TO_SHOW / 2);

  let start = Math.max(1, currentPage - half);
  let end = Math.min(totalPages, currentPage + half);

  if (currentPage - half <= 0) {
    end = Math.min(totalPages, MAX_PAGES_TO_SHOW);
  }

  if (currentPage + half > totalPages) {
    start = Math.max(1, totalPages - MAX_PAGES_TO_SHOW + 1);
  }

  for (let index = start; index <= end; index++) {
    pageNumbers.push(index);
  }

  if (start > 1) {
    if (start > 2) {
      pageNumbers.unshift(1, "ellipsis");
    } else {
      pageNumbers.unshift(1);
    }
  }

  if (end < totalPages - 1) {
    pageNumbers.push("ellipsis", totalPages);
  } else if (end === totalPages - 1) {
    pageNumbers.push(totalPages);
  }

  return pageNumbers;
}

export function getBudgetRenewalLabel(renewalType: BudgetRenewalType): string {
  switch (renewalType) {
    case "daily":
      return "day";
    case "weekly":
      return "week";
    case "monthly":
      return "month";
    case "yearly":
      return "year";
    case "never":
      return "never";
    case "":
      return "";
  }
}
