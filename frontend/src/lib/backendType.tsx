import { ReactElement } from "react";
import { LDKIcon } from "src/components/icons/LDK";
import { PhoenixdIcon } from "src/components/icons/Phoenixd";
import { BackendType } from "src/types";

import barkDark from "src/assets/images/node/bark-dark.svg";
import barkLight from "src/assets/images/node/bark-light.svg";
import cashu from "src/assets/images/node/cashu.png";
import cln from "src/assets/images/node/cln.png";
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
  LDK_SERVER: {
    title: "LDK Server",
    icon: <LDKIcon />,
    hasMnemonic: false,
    hasChannelManagement: true,
    hasNodeBackup: false,
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
