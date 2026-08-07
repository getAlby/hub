import { BackendType } from "src/types";

type BackendTypeConfig = {
  hasMnemonic: boolean;
  hasChannelManagement: boolean;
  hasNodeBackup: boolean;
};

export const backendTypeConfigs: Record<BackendType, BackendTypeConfig> = {
  LND: {
    hasMnemonic: false,
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
  LDK: {
    hasMnemonic: true,
    hasChannelManagement: true,
    hasNodeBackup: true,
  },
  PHOENIX: {
    hasMnemonic: false,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
  CASHU: {
    hasMnemonic: true,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
  CLN: {
    hasMnemonic: false,
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
  BARK: {
    hasMnemonic: true,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
  GREENLIGHT: {
    // 12-word BIP-39 is both GL node seed and Settings backup phrase
    hasMnemonic: true,
    // Real CLN channels on Blockstream cloud — not Bark stubs
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
};
