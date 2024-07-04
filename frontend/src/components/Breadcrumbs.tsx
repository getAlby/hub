import { Fragment } from "react";
import { Link, useMatches } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "src/components/ui/breadcrumb";

type MatchWithCrumb = {
  pathname: string;
  handle?: {
    crumb?: () => React.ReactNode;
  };
};

function Breadcrumbs() {
  const matches = useMatches() as MatchWithCrumb[]; // Type-cast useMatches result to MatchWithCrumb array

  const crumbs = matches
    // First, get rid of any matches that don't have a handle or crumb
    .filter(
      (
        match
      ): match is MatchWithCrumb & {
        handle: { crumb: () => React.ReactNode };
      } => Boolean(match.handle?.crumb)
    );

  // Compare pathnames of index routes to remove duplicates
  const isIndexRoute =
    crumbs.length >= 2 && crumbs[crumbs.length - 1].pathname
      ? crumbs[crumbs.length - 1].pathname.slice(0, -1) ===
        crumbs[crumbs.length - 2].pathname
      : false;

  // Remove the last item if it's an index route to prevent e.g. Wallet > Wallet
  const filteredCrumbs = isIndexRoute ? crumbs.slice(0, -1) : crumbs;

  // Don't render anything if there is only one item
  if (filteredCrumbs.length < 3) {
    return null;
  }

  return (
    <>
      <Breadcrumb>
        <BreadcrumbList>
          {filteredCrumbs.map((crumb, index) => (
            <Fragment key={index}>
              <BreadcrumbItem>
                {index + 1 < filteredCrumbs.length ? (
                  <BreadcrumbLink asChild>
                    <Link to={crumb.pathname}>{crumb.handle.crumb()}</Link>
                  </BreadcrumbLink>
                ) : (
                  <>{crumb.handle.crumb()}</>
                )}
              </BreadcrumbItem>
              {index + 1 < filteredCrumbs.length && <BreadcrumbSeparator />}
            </Fragment>
          ))}
        </BreadcrumbList>
      </Breadcrumb>
    </>
  );
}

export default Breadcrumbs;
