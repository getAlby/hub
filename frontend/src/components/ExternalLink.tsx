import { Link } from "react-router-dom";
import { openLink } from "src/utils/openLink";

type Props = {
  to: string;
  className?: string;
  children?: React.ReactNode;
};

export default function ExternalLink({ to, className, children }: Props) {
  const isHttpMode = window.location.protocol.startsWith("http");

  return isHttpMode ? (
    <Link
      to={to}
      target="_blank"
      rel="noreferer noopener"
      className={className}
    >
      {children}
    </Link>
  ) : (
    <div className={className} onClick={() => openLink(to)}>
      {children}
    </div>
  );
}
