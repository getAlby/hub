import { LucideIcon } from "lucide-react";
import { Button } from "src/components/ui/button.tsx";

type Props = {
  icon: LucideIcon;
  text: string;
  variant?: "outline";
};

const ResponsiveButton = ({ icon: Icon, text, variant }: Props) => {
  return (
    <>
      <Button className="hidden lg:inline-flex" variant={variant}>
        <Icon className="h-4 w-4 mr-2" />
        {text}
      </Button>
      <Button className="lg:hidden" variant={variant} size="icon">
        <Icon className="h-4 w-4" />
      </Button>
    </>
  );
};

export default ResponsiveButton;
