import React, { ReactElement } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { BreezIcon } from "src/components/icons/Breez";
import { GreenlightIcon } from "src/components/icons/Greenlight";
import { LDKIcon } from "src/components/icons/LDK";
import { PhoenixdIcon } from "src/components/icons/Phoenixd";
import { Button } from "src/components/ui/button";
import { backendTypeHasMnemonic, cn } from "src/lib/utils";
import { BackendType } from "src/types";

import cashu from "src/assets/images/node/cashu.png";
import lnd from "src/assets/images/node/lnd.png";
import useSetupStore from "src/state/SetupStore";

type BackendTypeDefinition = {
  id: BackendType;
  title: string;
  icon: ReactElement;
};

const backendTypes: BackendTypeDefinition[] = [
  {
    id: "LDK",
    title: "LDK",
    icon: <LDKIcon />,
  },
  {
    id: "PHOENIX",
    title: "phoenixd",
    icon: <PhoenixdIcon />,
  },
  {
    id: "BREEZ",
    title: "Breez SDK",
    icon: <BreezIcon />,
  },
  {
    id: "GREENLIGHT",
    title: "Greenlight",
    icon: <GreenlightIcon />,
  },
  {
    id: "LND",
    title: "LND",
    icon: <img src={lnd} />,
  },
  {
    id: "CASHU",
    title: "Cashu Mint",
    icon: <img src={cashu} />,
  },
];

export function SetupNode() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [selectedBackendType, setSelectedBackupType] =
    React.useState<BackendTypeDefinition>();

  function next() {
    navigate(`/setup/node/${selectedBackendType?.id.toLowerCase()}`);
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
            {backendTypes
              .filter((item) =>
                hasImportedMnemonic ? backendTypeHasMnemonic(item.id) : true
              )
              .map((item) => (
                <div
                  key={item.id}
                  className={cn(
                    "border-foreground-muted border px-4 py-6 flex flex-col gap-3 items-center rounded cursor-pointer",
                    selectedBackendType === item && "border-primary"
                  )}
                  onClick={() => setSelectedBackupType(item)}
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
