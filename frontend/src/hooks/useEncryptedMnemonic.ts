import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { MnemonicResponse } from "src/types";

export function useEncryptedMnemonic() {
  return useSWR<MnemonicResponse>("/api/mnemonic", swrFetcher);
}
