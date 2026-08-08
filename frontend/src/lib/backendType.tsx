import { ReactElement } from "react";
import { LDKIcon } from "src/components/icons/LDK";
import { PhoenixdIcon } from "src/components/icons/Phoenixd";
import { BackendType } from "src/types";

import barkDark from "src/assets/images/node/bark-dark.svg";
import barkLight from "src/assets/images/node/bark-light.svg";
import cashu from "src/assets/images/node/cashu.png";
import cln from "src/assets/images/node/cln.png";
import greenlight from "src/assets/images/node/greenlight-dark.svg";
import lnd from "src/assets/images/node/lnd.png";

type BackendTypeConfig = {
  title: string;
  icon: ReactElement;
  hasMnemonic: boolean;
  hasChannelManagement: boolean;
  hasNodeBackup: boolean;
};

export const backendTypeConfigs: Record<BackendType, BackendTypeConfig> = {
  LDK: {
    title: "LDK",
    icon: <LDKIcon />,
    hasMnemonic: true,
    hasChannelManagement: true,
    hasNodeBackup: true,
  },
  PHOENIX: {
    title: "phoenixd",
    icon: <PhoenixdIcon />,
    hasMnemonic: false,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
  LND: {
    title: "LND",
    icon: <img src={lnd} />,
    hasMnemonic: false,
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
  CASHU: {
    title: "Cashu Mint",
    icon: <img src={cashu} />,
    hasMnemonic: true,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
  CLN: {
    title: "CLN",
    icon: <img src={cln} />,
    hasMnemonic: false,
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
  GREENLIGHT: {
    title: "Greenlight",
    // the official Greenlight lockup is designed for dark backgrounds —
    // a black badge with the white emblem reads on the (always-light)
    // setup wizard panel AND on dark-mode surfaces; the white-badge
    // variant is invisible on the white wizard panel, so the dark badge
    // is used unconditionally (Bark's dark badge behaves the same)
    icon: <img src={greenlight} />,
    // 12-word BIP-39 is both GL node seed and Settings backup phrase
    hasMnemonic: true,
    // Real CLN channels on Blockstream cloud — not Bark stubs
    hasChannelManagement: true,
    hasNodeBackup: false,
  },
  BARK: {
    title: "Bark",
    icon: (
      <>
        <img src={barkLight} className="dark:hidden" />
        <img src={barkDark} className="hidden dark:block" />
      </>
    ),
    hasMnemonic: true,
    hasChannelManagement: false,
    hasNodeBackup: false,
  },
};
