import { Link } from "react-router-dom";
import { isHttpMode } from "src/utils/isHttpMode";
import { openLink } from "src/utils/openLink";

type Props = {
  to: string;
  className?: string;
  children?: React.ReactNode;
};

export default function ExternalLink({ to, className, children }: Props) {
  const _isHttpMode = isHttpMode();

  return _isHttpMode ? (
    <Link
      to={to}
      target="_blank"
      rel="noreferer noopener"
      className={className}
    >
      {children}
    </Link>
  ) : (
    <span className={className} onClick={() => openLink(to)}>
      {children}
    </span>
  );
}
