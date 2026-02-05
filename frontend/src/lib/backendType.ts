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
};
