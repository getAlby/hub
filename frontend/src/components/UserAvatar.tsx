import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { cn } from "src/lib/utils";

function UserAvatar({ className }: { className?: string }) {
  const { data: albyMe } = useAlbyMe();

  return (
    <Avatar className={cn("h-8 w-8 rounded-lg", className)}>
      <AvatarImage src={albyMe?.avatar} alt="Avatar" />
      <AvatarFallback className="font-medium rounded-lg">
        {(albyMe?.name || albyMe?.email || "SN").substring(0, 2).toUpperCase()}
      </AvatarFallback>
    </Avatar>
  );
}

export default UserAvatar;
