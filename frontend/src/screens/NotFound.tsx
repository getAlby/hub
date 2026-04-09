import { AlbyHubIcon } from "src/components/icons/AlbyHubIcon";
import { LinkButton } from "src/components/ui/custom/link-button";

function NotFound() {
  return (
    <>
      <title>Page Not Found · Alby Hub</title>
      <div className="flex flex-col items-center justify-center min-h-screen gap-6 text-center px-4">
        <AlbyHubIcon className="w-16 h-16 opacity-60" />
        <div>
          <h1 className="text-3xl font-semibold">Page not found</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            The page you're looking for doesn't exist
          </p>
        </div>
        <LinkButton to="/" variant="outline">
          Return Home
        </LinkButton>
      </div>
    </>
  );
}

export default NotFound;
