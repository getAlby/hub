export const NIP_47_PAY_INVOICE_METHOD = "pay_invoice"
export const NIP_47_GET_BALANCE_METHOD = "get_balance"
export const NIP_47_GET_INFO_METHOD = "get_info"
export const NIP_47_MAKE_INVOICE_METHOD = "make_invoice"
export const NIP_47_LOOKUP_INVOICE_METHOD = "lookup_invoice"

export type BackendType = "ALBY" | "LND"

export type RequestMethodType = "pay_invoice" | "get_balance" | "get_info" | "make_invoice" | "lookup_invoice";

export type BudgetRenewalType = "daily" | "weekly" | "monthly" | "yearly" | "";

export const validBudgetRenewals: BudgetRenewalType[] = ["daily", "weekly", "monthly", "yearly", ""]

export const nip47MethodDescriptions: Record<RequestMethodType, string> = {
	[NIP_47_GET_BALANCE_METHOD]: "Read your balance",
	[NIP_47_GET_INFO_METHOD]: "Read your node info",
	[NIP_47_PAY_INVOICE_METHOD]: "Send payments",
	[NIP_47_MAKE_INVOICE_METHOD]: "Create invoices",
	[NIP_47_LOOKUP_INVOICE_METHOD]: "Lookup status of invoices",
}

export const nip47MethodIcons: Record<RequestMethodType, string> = {
	[NIP_47_GET_BALANCE_METHOD]: "wallet",
	[NIP_47_GET_INFO_METHOD]: "wallet",
	[NIP_47_PAY_INVOICE_METHOD]: "lightning",
	[NIP_47_MAKE_INVOICE_METHOD]: "invoice",
	[NIP_47_LOOKUP_INVOICE_METHOD]: "search",
}

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

export interface UserInfo {
  user: User | null;
  backendType: BackendType;
  csrf: string;
}

export interface InfoResponse extends UserInfo{}

export interface ShowAppResponse {
  app: App;
  budgetUsage?: number;
  csrf: string;
  eventsCount: number;
  expiresAt?: number;
  expiresAtFormatted?: string;
  lastEvent?: NostrEvent;
  paySpecificPermission?: AppPermission;
  renewsIn?: string;
  requestMethods: RequestMethodType[];
}

export interface ListAppsResponse {
  apps: App[];
  lastEvent: Record<App["id"], NostrEvent>;
  eventsCounts: Record<App["id"], number>;
}
