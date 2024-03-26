import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { EncryptedMnemonicResponse } from "src/types";

export function useEncryptedMnemonic() {
  return useSWR<EncryptedMnemonicResponse>(
    "/api/encrypted-mnemonic",
    swrFetcher
  );
}
