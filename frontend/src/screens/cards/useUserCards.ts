import React from "react";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
} from "src/constants";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import type { App } from "src/types";

export type Provider = {
  id: string;
  name: string;
  logo: string;
  initials: string;
  network: "Visa" | "Mastercard";
  /**
   * If set, the provider has its own NWC pairing flow inside the in-hub app
   * store. The connect-card dialog routes the user to /appstore/<id> instead
   * of asking for a stablecoin top-up address.
   */
  appStoreId?: string;
};

// Mirrors bitcoin-card-topup's CardConfig (label/destinationAddress/chainId/currency).
// The NWC connection secret is *not* stored — it only lives in the one-shot
// CardCreatedDialog URL right after connection.
export type SupportedCurrency = "USDC" | "USDT";

export const CARD_APP_STORE_ID = "bitcoin-card-topup";

/** Metadata we attach to the NWC app so we can derive cards from /api/apps. */
type CardMetadata = {
  app_store_app_id: string;
  card_provider_id?: string;
  card_destination_address?: string;
  card_chain_id?: number;
  card_currency?: SupportedCurrency;
};

export type UserCard = {
  /** id of the NWC app that backs this card. */
  appId: number;
  providerId: string;
  destinationAddress: string;
  chainId: number;
  currency: SupportedCurrency;
  createdAt: string;
};

function appToCard(app: App): UserCard | null {
  const m = app.metadata as CardMetadata | undefined;
  if (!m) {
    return null;
  }
  if (
    !m.card_provider_id ||
    !m.card_destination_address ||
    typeof m.card_chain_id !== "number" ||
    !m.card_currency
  ) {
    return null;
  }
  return {
    appId: app.id,
    providerId: m.card_provider_id,
    destinationAddress: m.card_destination_address,
    chainId: m.card_chain_id,
    currency: m.card_currency,
    createdAt: app.createdAt,
  };
}

type AddCardInput = {
  providerId: string;
  providerName: string;
  destinationAddress: string;
  chainId: number;
  currency: SupportedCurrency;
};

type AddCardResult = { card: UserCard; pairingUri: string };

export function useUserCards() {
  const { data, mutate } = useApps(
    100,
    1,
    { appStoreAppId: CARD_APP_STORE_ID },
    "created_at"
  );

  const cards = data?.apps
    ? data.apps.map(appToCard).filter((c): c is UserCard => c !== null)
    : [];

  const addCard = React.useCallback(
    async (input: AddCardInput): Promise<AddCardResult> => {
      // Mint a real NWC connection — this is the one moment we hold the
      // pairing URI; after this it's gone (write-only secret).
      const metadata: CardMetadata = {
        app_store_app_id: CARD_APP_STORE_ID,
        card_provider_id: input.providerId,
        card_destination_address: input.destinationAddress,
        card_chain_id: input.chainId,
        card_currency: input.currency,
      };

      const response = await createApp({
        name: `${input.providerName} Card`,
        scopes: [
          "get_info",
          "get_balance",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "pay_invoice",
          "notifications",
        ],
        maxAmountSat: DEFAULT_APP_BUDGET_SATS,
        budgetRenewal: DEFAULT_APP_BUDGET_RENEWAL,
        metadata,
      });

      const card: UserCard = {
        appId: response.id,
        providerId: input.providerId,
        destinationAddress: input.destinationAddress,
        chainId: input.chainId,
        currency: input.currency,
        createdAt: new Date().toISOString(),
      };

      // Refresh the apps list so the new card appears.
      mutate();

      return { card, pairingUri: response.pairingUri };
    },
    [mutate]
  );

  return { cards, addCard };
}
