import { VariantProps } from "class-variance-authority";
import * as React from "react";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { buttonVariants } from "./ui/buttonVariants";

type Props = {
  icon: React.ComponentType;
  text: string;
  to: string;
} & React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

const ResponsiveExternalLinkButton = ({
  icon: Icon,
  text,
  variant,
  ...props
}: Props) => {
  return (
    <>
      <ExternalLinkButton
        {...props}
        className="hidden lg:inline-flex"
        variant={variant}
      >
        <Icon />
        {text}
      </ExternalLinkButton>
      <ExternalLinkButton
        {...props}
        size="icon"
        className="lg:hidden"
        variant={variant}
      >
        <Icon />
      </ExternalLinkButton>
    </>
  );
};

export default ResponsiveExternalLinkButton;
