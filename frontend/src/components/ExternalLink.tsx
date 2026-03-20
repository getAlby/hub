import { Link } from "react-router-dom";
import { isHttpMode } from "src/utils/isHttpMode";
import { openLink } from "src/utils/openLink";

type Props = React.PropsWithChildren<{
  to: string;
  className?: string;
}> &
  Omit<
    React.AnchorHTMLAttributes<HTMLAnchorElement>,
    "href" | "className" | "children"
  >;

export default function ExternalLink({
  to,
  className,
  children,
  ...rest
}: Props) {
  const _isHttpMode = isHttpMode();

  return _isHttpMode ? (
    <Link
      to={to}
      target="_blank"
      rel="noreferer noopener"
      className={className}
      {...rest}
    >
      {children}
    </Link>
  ) : (
    <span className={className} onClick={() => openLink(to)} {...rest}>
      {children}
    </span>
  );
}
