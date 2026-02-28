import React, { ReactElement } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LDKIcon } from "src/components/icons/LDK";
import { PhoenixdIcon } from "src/components/icons/Phoenixd";
import { Button } from "src/components/ui/button";
import { cn } from "src/lib/utils";
import { BackendType } from "src/types";

import cashu from "src/assets/images/node/cashu.png";
import lnd from "src/assets/images/node/lnd.png";
import { backendTypeConfigs } from "src/lib/backendType";
import useSetupStore from "src/state/SetupStore";

type BackendTypeDisplayConfig = {
  title: string;
  icon: ReactElement;
};

const backendTypeDisplayConfigs: Partial<
  Record<BackendType, BackendTypeDisplayConfig>
> = {
  LDK: {
    title: "LDK",
    icon: <LDKIcon />,
  },
  PHOENIX: {
    title: "phoenixd",
    icon: <PhoenixdIcon />,
  },
  LND: {
    title: "LND",
    icon: <img src={lnd} />,
  },
  CASHU: {
    title: "Cashu Mint",
    icon: <img src={cashu} />,
  },
  BARK: {
    title: "Bark",
    icon: <LDKIcon />, // FIXME: proper icon
  },
};

const backendTypeDisplayConfigList = Object.entries(
  backendTypeDisplayConfigs
).map((entry) => ({
  ...entry[1],
  backendType: entry[0] as BackendType,
}));

export function SetupNode() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [selectedBackendType, setSelectedBackupType] =
    React.useState<BackendType>();

  function next() {
    if (!selectedBackendType) {
      return;
    }
    navigate(`/setup/node/${selectedBackendType.toLowerCase()}`);
  }

  const hasImportedMnemonic = !!setupStore.nodeInfo.mnemonic;

  return (
    <>
      <Container>
        <TwoColumnLayoutHeader
          title="Choose Wallet Implementation"
          description="Decide between one of available lightning wallet backends."
        />
        <div className="flex flex-col gap-5 w-full mt-6">
          <div className="w-full grid grid-cols-2 gap-4">
            {backendTypeDisplayConfigList
              .filter((item) =>
                hasImportedMnemonic
                  ? backendTypeConfigs[item.backendType].hasMnemonic
                  : true
              )
              .map((item) => (
                <div
                  key={item.backendType}
                  className={cn(
                    "border-foreground-muted border px-4 py-6 flex flex-col gap-3 items-center rounded cursor-pointer",
                    selectedBackendType === item.backendType && "border-primary"
                  )}
                  onClick={() => setSelectedBackupType(item.backendType)}
                >
                  <div className="h-6 w-6">{item.icon}</div>
                  {item.title}
                </div>
              ))}
          </div>
          <Button onClick={() => next()} disabled={!selectedBackendType}>
            Next
          </Button>
        </div>
      </Container>
    </>
  );
}
