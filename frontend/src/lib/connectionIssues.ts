import { ConnectionIssue } from "src/types";

export type ConnectionIssueCopy = {
  title: string;
  description: string;
  action: string;
  href?: string;
  onClick?: () => void;
};

export function getMethodLabel(method: string) {
  switch (method) {
    case "pay_invoice":
    case "multi_pay_invoice":
    case "pay_keysend":
    case "multi_pay_keysend":
      return "send payments";
    case "make_invoice":
      return "create invoices";
    case "lookup_invoice":
      return "look up invoices";
    case "list_transactions":
      return "read transaction history";
    case "get_balance":
      return "read the balance";
    case "get_info":
      return "read wallet info";
    case "sign_message":
      return "sign messages";
    default:
      return "use this wallet feature";
  }
}

export function getConnectionIssueCopy(
  appName: string,
  issue: ConnectionIssue,
  onViewDetails: () => void
): ConnectionIssueCopy {
  switch (issue.category) {
    case "missing_permission":
      return {
        title: "App needs permission",
        description: `This connection does not allow ${getMethodLabel(issue.method)} yet.`,
        action: "Review connection",
        href: `/apps/${issue.appId}?edit`,
      };
    case "unknown_method":
      return {
        title: "App requested an unknown feature",
        description:
          "Alby Hub does not recognize this app request. No wallet action was taken.",
        action: "View details",
        onClick: onViewDetails,
      };
    case "expired_connection":
      return {
        title: "Connection expired",
        description:
          "This app connection has expired. Review the connection to renew access.",
        action: "Review connection",
        href: `/apps/${issue.appId}?edit`,
      };
    case "budget_exceeded":
      return {
        title: "Connection budget reached",
        description:
          "This payment is above the connection's remaining budget. No payment was sent.",
        action: "Review budget",
        href: `/apps/${issue.appId}?edit`,
      };
    case "low_balance":
      return {
        title: "Not enough spendable balance",
        description:
          "This wallet does not have enough spendable sats for the app request.",
        action: "View details",
        onClick: onViewDetails,
      };
    case "payment_failed":
      return {
        title: "Payment was not sent",
        description: `Alby Hub could not complete the payment requested by ${appName}. Check transactions if you are unsure.`,
        action: "View details",
        onClick: onViewDetails,
      };
    default:
      return {
        title: `${appName} request failed`,
        description:
          "Alby Hub could not complete this app request. No wallet action was taken unless you see it in your transactions.",
        action: "View details",
        onClick: onViewDetails,
      };
  }
}
