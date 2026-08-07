import { wordlist } from "@scure/bip39/wordlists/english.js";
import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import useSetupStore from "src/state/SetupStore";

import * as bip39 from "@scure/bip39";
import Loading from "src/components/Loading";

// Greenlight uses the same 12-word BIP-39 phrase as gl-cli
// (mnemonic.to_seed("")[0..32]). Strength 128 → 12 words.
export function GreenlightForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    if (!useSetupStore.getState().nodeInfo.mnemonic) {
      useSetupStore.getState().updateNodeInfo({
        mnemonic: bip39.generateMnemonic(wordlist, 128),
      });
    }
    useSetupStore.getState().updateNodeInfo({
      backendType: "GREENLIGHT",
    });
    navigate("/setup/security", {
      replace: true,
    });
  }, [navigate, searchParams]);

  return (
    <>
      <title>Loading... · Alby Hub</title>
      <Loading />
    </>
  );
}
