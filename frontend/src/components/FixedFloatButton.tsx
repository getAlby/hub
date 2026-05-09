import * as React from "react";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";

type FixedFloatButtonProps = Omit<
  React.ComponentProps<typeof ExternalLinkButton>,
  "to"
> & {
  address?: string;
  from?: string;
  to?: string;
};

export function FixedFloatButton({
  address,
  from,
  to,
  ...props
}: FixedFloatButtonProps) {
  const params = new URLSearchParams({
    ref: "qnnjvywb",
  });
  if (from) {
    params.set("from", from);
  }
  if (to) {
    params.set("to", to);
  }
  if (address) {
    params.set("address", address);
  }

  return (
    <ExternalLinkButton to={`https://ff.io/?${params.toString()}`} {...props} />
  );
}
