import { wordlist } from "@scure/bip39/wordlists/english.js";
import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import * as bip39 from "@scure/bip39";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";
import { backendTypeConfigs } from "src/lib/backendType";

export function PresetNodeForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { data: info } = useInfo();

  // No configuration needed, automatically proceed with the next step
  useEffect(() => {
    if (!info) {
      return;
    }
    if (backendTypeConfigs[info.backendType].hasMnemonic) {
      useSetupStore.getState().updateNodeInfo({
        mnemonic: bip39.generateMnemonic(wordlist, 128),
      });
    }

    navigate("/setup/security");
  }, [info, navigate, searchParams]);

  return <Loading />;
}
