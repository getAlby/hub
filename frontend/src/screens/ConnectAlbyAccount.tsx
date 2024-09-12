import { LinkButton } from "src/components/ui/button";

export function ConnectAlbyAccount() {
  return (
    <div className="w-full h-full flex flex-col items-center justify-center gap-5">
      <h1 className="text-2xl font-semibold">Connect Your Alby Account</h1>
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
      <p className="max-w-sm md:max-w-lg text-muted-foreground text-center">
        Your Alby Account gives your hub a lightning address, Nostr address and
        zaps, email notifications, fiat topups, priority support, automatic
        channel backups, access to podcasting apps & more.
      </p>
      <div className="flex flex-col items-center justify-center mt-5 gap-5">
        <LinkButton to="/alby/auth" size="lg">
          Connect now
        </LinkButton>
        <LinkButton variant="link" to="/" className="text-muted-foreground">
          Maybe later
        </LinkButton>
      </div>
    </div>
  );
}
