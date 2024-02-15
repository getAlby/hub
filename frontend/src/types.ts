export const NIP_47_PAY_INVOICE_METHOD = "pay_invoice";
export const NIP_47_GET_BALANCE_METHOD = "get_balance";
export const NIP_47_GET_INFO_METHOD = "get_info";
export const NIP_47_MAKE_INVOICE_METHOD = "make_invoice";
export const NIP_47_LOOKUP_INVOICE_METHOD = "lookup_invoice";
export const NIP_47_LIST_TRANSACTIONS_METHOD = "list_transactions";

export type BackendType = "LND" | "BREEZ" | "GREENLIGHT";

export type RequestMethodType =
  | "pay_invoice"
  | "get_balance"
  | "get_info"
  | "make_invoice"
  | "lookup_invoice"
  | "list_transactions";

export type BudgetRenewalType = "daily" | "weekly" | "monthly" | "yearly" | "";

export type IconMap = {
  [key in RequestMethodType]: (
    props: React.SVGAttributes<SVGElement>
  ) => JSX.Element;
};

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
  [NIP_47_LIST_TRANSACTIONS_METHOD]: "Read incoming transaction history",
  [NIP_47_LOOKUP_INVOICE_METHOD]: "Lookup status of invoices",
  [NIP_47_MAKE_INVOICE_METHOD]: "Create invoices",
  [NIP_47_PAY_INVOICE_METHOD]: "Send payments",
};

export interface ErrorResponse {
  message: string;
}

export interface App {
  id: number;
  userId: number;
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

// export interface AppPermission {
//   id: number;
//   appId: number;
//   app: App;
//   requestMethod: RequestMethodType;
//   maxAmount: number;
//   budgetRenewal: string;
//   expiresAt: string;
//   createdAt: string;
//   updatedAt: string;
// }

export interface InfoResponse {
  backendType: BackendType;
  setupCompleted: boolean;
  running: boolean;
  unlocked: boolean;
}

export interface CreateAppResponse {
  name: string;
  pairingUri: string;
  pairingPublicKey: string;
  pairingSecretKey: string;
  returnTo: string;
}

export type Channel = {
  localBalance: number;
  remoteBalance: number;
  remotePubkey: string;
  id: string;
  active: boolean;
};

export type NodeConnectionInfo = {
  pubkey: string;
  address: string;
  port: number;
};

export type ConnectPeerRequest = {
  pubkey: string;
  address: string;
  port: number;
};

export type OpenChannelRequest = {
  pubkey: string;
  amount: number;
};

export type OpenChannelResponse = {
  fundingTxId: string;
};

export type GetOnchainAddressResponse = {
  address: string;
};

export type OnchainBalanceResponse = {
  sats: number;
};

// from https://mempool.space/docs/api/rest#get-node-stats
export type Node = {
  alias: string;
  public_key: string;
  color: string;
  active_channel_count: number;
  sockets: string;
};
export type NodeInfo = Partial<{
  backendType: BackendType;

  mnemonic?: string;
  greenlightInviteCode?: string;
  breezApiKey?: string;

  lndAddress?: string;
  lndCertHex?: string;
  lndMacaroonHex?: string;
}>;
