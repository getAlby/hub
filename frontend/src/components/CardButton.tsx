import { ChevronRightIcon } from "lucide-react";
import { ReactElement } from "react";
import { Link } from "react-router-dom";
import { Card } from "src/components/ui/card";

type Props = {
  title: string | ReactElement;
  description: string | ReactElement;
  to: string;
};

function CardButton({ title, description, to }: Props) {
  return (
    <Link to={to}>
      <Card className="p-4 shadow-none hover:bg-accent">
        <div className="flex flex-row justify-between items-center">
          <div>
            <div className="font-medium flex flex-row items-center gap-2">
              {title}
            </div>
            <div className="text-muted-foreground text-sm">{description}</div>
          </div>
          <div>
            <ChevronRightIcon />
          </div>
        </div>
      </Card>
    </Link>
  );
}

export default CardButton;
