import {
  Bell,
  CirclePlus,
  HandCoins,
  Info,
  LucideIcon,
  NotebookTabs,
  PenLine,
  Search,
  WalletMinimal,
} from "lucide-react";

export const NIP_47_PAY_INVOICE_METHOD = "pay_invoice";
export const NIP_47_GET_BALANCE_METHOD = "get_balance";
export const NIP_47_GET_INFO_METHOD = "get_info";
export const NIP_47_MAKE_INVOICE_METHOD = "make_invoice";
export const NIP_47_LOOKUP_INVOICE_METHOD = "lookup_invoice";
export const NIP_47_LIST_TRANSACTIONS_METHOD = "list_transactions";
export const NIP_47_SIGN_MESSAGE_METHOD = "sign_message";

export const NIP_47_NOTIFICATIONS_PERMISSION = "notifications";

export type BackendType =
  | "LND"
  | "BREEZ"
  | "GREENLIGHT"
  | "LDK"
  | "PHOENIX"
  | "CASHU";

export type RequestMethodType =
  | "pay_invoice"
  | "get_balance"
  | "get_info"
  | "make_invoice"
  | "lookup_invoice"
  | "list_transactions"
  | "sign_message";

export type BudgetRenewalType =
  | "daily"
  | "weekly"
  | "monthly"
  | "yearly"
  | "never"
  | "";

// TODO: move other permissions
export type PermissionType =
  | RequestMethodType
  | typeof NIP_47_NOTIFICATIONS_PERMISSION;

export type IconMap = {
  [key in PermissionType]: LucideIcon;
};

export const iconMap: IconMap = {
  [NIP_47_GET_BALANCE_METHOD]: WalletMinimal,
  [NIP_47_GET_INFO_METHOD]: Info,
  [NIP_47_LIST_TRANSACTIONS_METHOD]: NotebookTabs,
  [NIP_47_LOOKUP_INVOICE_METHOD]: Search,
  [NIP_47_MAKE_INVOICE_METHOD]: CirclePlus,
  [NIP_47_PAY_INVOICE_METHOD]: HandCoins,
  [NIP_47_SIGN_MESSAGE_METHOD]: PenLine,
  [NIP_47_NOTIFICATIONS_PERMISSION]: Bell,
};

export const validBudgetRenewals: BudgetRenewalType[] = [
  "daily",
  "weekly",
  "monthly",
  "yearly",
  "never",
];

export const nip47MethodDescriptions: Record<RequestMethodType, string> = {
  [NIP_47_GET_BALANCE_METHOD]: "Read your balance",
  [NIP_47_GET_INFO_METHOD]: "Read your node info",
  [NIP_47_LIST_TRANSACTIONS_METHOD]: "Read incoming transaction history",
  [NIP_47_LOOKUP_INVOICE_METHOD]: "Lookup status of invoices",
  [NIP_47_MAKE_INVOICE_METHOD]: "Create invoices",
  [NIP_47_PAY_INVOICE_METHOD]: "Send payments",
  [NIP_47_SIGN_MESSAGE_METHOD]: "Sign messages",
};

// TODO: merge with nip47MethodDescriptions
export const nip47PermissionDescriptions: Record<PermissionType, string> = {
  ...nip47MethodDescriptions,
  [NIP_47_NOTIFICATIONS_PERMISSION]: "Receive wallet notifications",
};

export const expiryOptions: Record<string, number> = {
  "1 week": 7,
  "1 month": 30,
  "1 year": 365,
  Never: 0,
};

export const budgetOptions: Record<string, number> = {
  "10k": 10_000,
  "25k": 25_000,
  "50k": 50_000,
  "100k": 100_000,
  "1M": 1_000_000,
  Unlimited: 0,
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

  // TODO: rename
  requestMethods: PermissionType[];
  maxAmount: number;
  budgetUsage: number;
  budgetRenewal: string;
}

export interface AppPermissions {
  // TODO: rename to permissions
  requestMethods: Set<PermissionType>;
  maxAmount: number;
  budgetRenewal: BudgetRenewalType;
  expiresAt?: Date;
}

export interface InfoResponse {
  backendType: BackendType;
  setupCompleted: boolean;
  oauthRedirect: boolean;
  albyAccountConnected: boolean;
  running: boolean;
  unlocked: boolean;
  albyAuthUrl: string;
  nextBackupReminder: string;
  albyUserIdentifier: string;
  network?: Network;
  version: string;
  latestVersion: string;
}

export type Network = "bitcoin" | "testnet" | "signet";

export interface EncryptedMnemonicResponse {
  mnemonic: string;
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
  localSpendableBalance: number;
  remoteBalance: number;
  remotePubkey: string;
  id: string;
  fundingTxId: string;
  active: boolean;
  public: boolean;
  confirmations?: number;
  confirmationsRequired?: number;
  forwardingFeeBaseMsat: number;
  unspendablePunishmentReserve: number;
  counterpartyUnspendablePunishmentReserve: number;
};

export type UpdateChannelRequest = {
  forwardingFeeBaseMsat: number;
};

export type Peer = {
  nodeId: string;
  address: string;
  isPersisted: boolean;
  isConnected: boolean;
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

export type SignMessageRequest = {
  message: string;
};

export type SignMessageResponse = {
  message: string;
  signature: string;
};

export type PayInvoiceResponse = {
  preimage: string;
  fee: number;
};

export type CreateInvoiceRequest = {
  amount: number;
  description: string;
};

export type OpenChannelRequest = {
  pubkey: string;
  amount: number;
  public: boolean;
};

export type OpenChannelResponse = {
  fundingTxId: string;
};

// eslint-disable-next-line @typescript-eslint/ban-types
export type CloseChannelResponse = {};

export type OnchainBalanceResponse = {
  spendable: number;
  total: number;
  reserved: number;
};

// from https://mempool.space/docs/api/rest#get-node-stats
export type Node = {
  alias: string;
  public_key: string;
  color: string;
  active_channel_count: number;
  sockets: string;
};
export type SetupNodeInfo = Partial<{
  backendType: BackendType;

  mnemonic?: string;
  nextBackupReminder?: string;
  greenlightInviteCode?: string;
  breezApiKey?: string;

  lndAddress?: string;
  lndCertHex?: string;
  lndMacaroonHex?: string;

  phoenixdAddress?: string;
  phoenixdAuthorization?: string;
}>;

export type LSPType = "LSPS1" | "Flow 2.0" | "PMLSP";

export type RecommendedChannelPeer = {
  network: Network;
  image: string;
  name: string;
  minimumChannelSize: number;
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      lspType: LSPType;
      lspUrl: string;
      pubkey?: string;
    }
);

// TODO: move to different file
export type AlbyMe = {
  identifier: string;
  nostr_pubkey: string;
  lightning_address: string;
  email: string;
  name: string;
  avatar: string;
  keysend_pubkey: string;
  shared_node: boolean;
};

export type AlbyBalance = {
  sats: number;
};

export type NewInstantChannelInvoiceRequest = {
  amount: number;
  lspType: LSPType;
  lspUrl: string;
  public: boolean;
};

export type NewInstantChannelInvoiceResponse = {
  invoice: string;
  fee: number;
  invoiceAmount: number;
  incomingLiquidity: number;
  outgoingLiquidity: number;
};

export type RedeemOnchainFundsResponse = {
  txId: string;
};

export type LightningBalanceResponse = {
  totalSpendable: number;
  totalReceivable: number;
  nextMaxSpendable: number;
  nextMaxReceivable: number;
  nextMaxSpendableMPP: number;
  nextMaxReceivableMPP: number;
};

export type BalancesResponse = {
  onchain: OnchainBalanceResponse;
  lightning: LightningBalanceResponse;
};

export type Transaction = {
  type: string;
  invoice: string;
  description: string;
  description_hash: string;
  preimage: string;
  payment_hash: string;
  amount: number;
  fees_paid: number;
  created_at: number;
  expires_at: number;
  settled_at: number;
  metadata: string[];
};

export type NewChannelOrderStatus = "pay" | "success" | "opening";

export type NewChannelOrder = {
  amount: string;
  isPublic: boolean;
  status: NewChannelOrderStatus;
  fundingTxId?: string;
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      lspType: LSPType;
      lspUrl: string;
    }
);
