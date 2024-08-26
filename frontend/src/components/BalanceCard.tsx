import { LucideIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

type BalanceCardProps = {
  title: string;
  balance: string;
  buttonTitle: string;
  buttonLink: string;
  hasChannelManagement: boolean;
  BalanceCardIcon: LucideIcon;
};

function BalanceCard({
  title,
  balance,
  buttonTitle,
  buttonLink,
  hasChannelManagement,
  BalanceCardIcon,
}: BalanceCardProps) {
  return (
    <Card className="w-full hidden md:block self-start">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <BalanceCardIcon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {!balance && (
          <div>
            <div className="animate-pulse d-inline ">
              <div className="h-2.5 bg-primary rounded-full w-12 my-2"></div>
            </div>
          </div>
        )}
        {balance && (
          <div className="text-2xl font-bold balance sensitive">
            {balance} sats
          </div>
        )}
      </CardContent>
      {hasChannelManagement && (
        <CardFooter className="flex justify-end">
          <Link to={buttonLink}>
            <Button variant="outline">{buttonTitle}</Button>
          </Link>
        </CardFooter>
      )}
    </Card>
  );
}

export default BalanceCard;
