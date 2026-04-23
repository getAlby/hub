import { Link } from "react-router";
import { isHttpMode } from "src/utils/isHttpMode";
import { openLink } from "src/utils/openLink";

type Props = React.HTMLAttributes<HTMLElement> & {
  to: string;
  children?: React.ReactNode;
};

export default function ExternalLink({ to, children, ...props }: Props) {
  const _isHttpMode = isHttpMode();

  return _isHttpMode ? (
    <Link to={to} target="_blank" rel="noreferer noopener" {...props}>
      {children}
    </Link>
  ) : (
    <span {...props} onClick={() => openLink(to)}>
      {children}
    </span>
  );
}
