import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function CashuForm() {
  const setupStore = useSetupStore();
  const navigate = useNavigate();
  const [cashuMintUrl, setCashuMintUrl] = React.useState(
    "https://mint.minibits.cash/Bitcoin"
  );

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({ cashuMintUrl });
  }

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "CASHU",
      ...data,
    });
    navigate("/setup/security");
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure Cashu Mint"
        description="Fill out wallet details to finish setup."
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
        <div className="grid gap-1.5">
          <Label
            htmlFor="cashu-mint-url"
            className="flex flex-row justify-between"
          >
            <span>Cashu Mint URL</span>{" "}
            <ExternalLink
              to="https://bitcoinmints.com"
              className="underline hover:no-underline text-xs font-normal"
            >
              Find a mint
            </ExternalLink>
          </Label>
          <Input
            name="cashu-mint-url"
            onChange={(e) => setCashuMintUrl(e.target.value)}
            value={cashuMintUrl}
            id="cashu-mint-url"
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
