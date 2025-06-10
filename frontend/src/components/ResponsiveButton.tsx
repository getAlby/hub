import { LucideIcon } from "lucide-react";
import { Button, ButtonProps } from "src/components/ui/button.tsx";

type Props = {
  icon: LucideIcon;
  text: string;
};

const ResponsiveButton = ({
  icon: Icon,
  text,
  variant,
  ...props
}: Props & ButtonProps) => {
  return (
    <>
      <Button {...props} className="hidden lg:inline-flex" variant={variant}>
        <Icon className="h-4 w-4 mr-2" />
        {text}
      </Button>
      <Button {...props} size="icon" className="lg:hidden" variant={variant}>
        <Icon className="h-4 w-4" />
      </Button>
    </>
  );
};

export default ResponsiveButton;
