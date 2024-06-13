import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";

function UserAvatar({ className }: { className?: string }) {
  const { data: albyMe } = useAlbyMe();

  return (
    <Avatar className={className}>
      <AvatarImage src={albyMe?.avatar} alt="Avatar" />
      <AvatarFallback>
        {(albyMe?.name || albyMe?.email || "SN").substring(0, 2).toUpperCase()}
      </AvatarFallback>
    </Avatar>
  );
}

export default UserAvatar;
