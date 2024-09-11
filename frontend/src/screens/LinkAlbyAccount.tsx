import { LinkButton } from "src/components/ui/button";

export function LinkAlbyAccount() {
  return (
    <div className="w-full h-full min-h-screen flex flex-col items-center justify-center gap-5">
      <div className="w-full max-w-screen-sm">
        <img
          src="/images/illustrations/alby-account-dark.svg"
          className="w-full hidden dark:block"
        />
        <img
          src="/images/illustrations/alby-account-light.svg"
          className="w-full dark:hidden"
        />
      </div>
      <p className="max-w-md text-muted-foreground text-center">
        Your Alby Account gives your hub a lightning address, Nostr address and
        zaps, email notifications, automatic channel backups, access to
        podcasting apps & more.
      </p>
      <div className="flex flex-col items-center justify-center mt-5 gap-2">
        <LinkButton to="/alby/auth">Connect now</LinkButton>
        <LinkButton variant="ghost" to="/">
          Maybe later
        </LinkButton>
      </div>
    </div>
  );
}
