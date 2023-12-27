export const NIP_47_PAY_INVOICE_METHOD = "pay_invoice";
export const NIP_47_GET_BALANCE_METHOD = "get_balance";
export const NIP_47_GET_INFO_METHOD = "get_info";
export const NIP_47_MAKE_INVOICE_METHOD = "make_invoice";
export const NIP_47_LOOKUP_INVOICE_METHOD = "lookup_invoice";
export const NIP_47_LIST_TRANSACTIONS_METHOD = "list_transactions";

export type BackendType = "ALBY" | "LND";

export type RequestMethodType =
  | "pay_invoice"
  | "get_balance"
  | "get_info"
  | "make_invoice"
  | "lookup_invoice"
  | "list_transactions";

export type BudgetRenewalType = "daily" | "weekly" | "monthly" | "yearly" | "";

export const validBudgetRenewals: BudgetRenewalType[] = [
  "daily",
  "weekly",
  "monthly",
  "yearly",
  "",
];

export const nip47MethodDescriptions: Record<RequestMethodType, string> = {
  [NIP_47_GET_BALANCE_METHOD]: "Read your balance",
  [NIP_47_GET_INFO_METHOD]: "Read your node info",
  [NIP_47_PAY_INVOICE_METHOD]: "Send payments",
  [NIP_47_MAKE_INVOICE_METHOD]: "Create invoices",
  [NIP_47_LOOKUP_INVOICE_METHOD]: "Lookup status of invoices",
  [NIP_47_LIST_TRANSACTIONS_METHOD]: "Read incoming transaction history",
};

export const nip47MethodIcons: Record<RequestMethodType, string> = {
  [NIP_47_GET_BALANCE_METHOD]: "wallet",
  [NIP_47_GET_INFO_METHOD]: "wallet",
  [NIP_47_PAY_INVOICE_METHOD]: "lightning",
  [NIP_47_MAKE_INVOICE_METHOD]: "invoice",
  [NIP_47_LOOKUP_INVOICE_METHOD]: "search",
  [NIP_47_LIST_TRANSACTIONS_METHOD]: "transactions",
};

export interface User {
  id: number;
  albyIdentifier: string;
  accessToken: string;
  refreshToken: string;
  email: string;
  expiry: string;
  lightningAddress: string;
  apps: string;
  createdAt: string;
  updatedAt: string;
  expiresAt: string;
}

export interface App {
  id: number;
  userId: number;
  user: User;
  name: string;
  description: string;
  nostrPubkey: string;
  createdAt: string;
  updatedAt: string;
  lastEventAt?: string;
  expiresAt?: string;

  requestMethods: string[];
  maxAmount: number;
  budgetUsage: number;
  budgetRenewal: string;
}

export interface AppPermission {
  id: number;
  appId: number;
  app: App;
  requestMethod: RequestMethodType;
  maxAmount: number;
  budgetRenewal: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface NostrEvent {
  id: number;
  appId: number;
  app: App;
  nostrId: string;
  replyId: string;
  content: string;
  state: string;
  repliedAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface InfoResponse {
  backendType: BackendType;
}

export interface CreateAppResponse {
  pairingUri: string;
}
