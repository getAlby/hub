import { wordlist } from "@scure/bip39/wordlists/english.js";
import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import * as bip39 from "@scure/bip39";
import Loading from "src/components/Loading";

export function LDKForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  // No configuration needed, automatically proceed with the next step
  useEffect(() => {
    // only generate a mnemonic if one is not already imported
    if (!useSetupStore.getState().nodeInfo.mnemonic) {
      useSetupStore.getState().updateNodeInfo({
        mnemonic: bip39.generateMnemonic(wordlist, 128),
      });
    }
    useSetupStore.getState().updateNodeInfo({
      backendType: "LDK",
    });
    navigate("/setup/security");
  }, [navigate, searchParams]);

  return <Loading />;
}
