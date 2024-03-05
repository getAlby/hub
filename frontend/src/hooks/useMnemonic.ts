import useSWR from "swr";

import { swrFetcher } from "src/utils/swr";
import { MnemonicResponse } from "src/types";

export function useMnemonic() {
  return useSWR<MnemonicResponse>("/api/mnemonic", swrFetcher);
}
