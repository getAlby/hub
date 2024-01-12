import nostrBitcoin from "src/assets/images/nostr-bitcoin.png";

function About() {
  return (
    <>
      <div className="bg-white rounded-md shadow p-4 lg:p-8 mb-8 dark:bg-surface-02dp">
        <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">
          tl;dr
        </h3>
        <p className="leading-relaxed dark:text-neutral-400">
          Nostr Wallet Connect (NWC) is a way for applications like Nostr
          clients to access a remote Lightning wallet through a{" "}
          <a
            href="https://github.com/getAlby/nips/blob/master/47.md"
            className="dark:text-gray-300 underline"
          >
            standardized protocol
          </a>
          . For instance, you can link your{" "}
          <a
            href="https://getalby.com"
            className="dark:text-gray-300 underline"
          >
            Alby account
          </a>{" "}
          with any other application using this feature.
        </p>
      </div>

      <div className="bg-white rounded-md shadow p-4 lg:p-8 mb-8 dark:bg-surface-02dp">
        <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">
          How to connect your app?
        </h3>
        <ol className="list-decimal list-inside leading-relaxed dark:text-neutral-400">
          <li className="mb-2">
            <strong className="dark:text-white">
              Create a new app connection
            </strong>
            <p className="ml-6">
              Enable the application by creating a permissioned connection to
              your wallet. This access can be revoked at any time.
            </p>
          </li>
          <li className="mb-2">
            <strong className="dark:text-white">Connect your app</strong>
            <p className="ml-6">
              Connect your application by adding the connection secret. Either
              by scanning the connect QR code or clicking a link.
            </p>
          </li>
          <li className="mb-2">
            <strong className="dark:text-white">Done</strong>
            <p className="ml-6">
              Zapp and Boost! Your lightning wallet is available where you need
              it.
              <br />
              And if you ever want to disable it again, simply disconnect the
              connection again.
            </p>
          </li>
        </ol>
      </div>

      <div className="bg-white rounded-md shadow p-4 lg:p-8 mb-8 dark:bg-surface-02dp">
        <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">
          FAQs
        </h3>
        <h4 className="font-bold text-l font-headline mt-4 dark:text-white">
          How does it work?
        </h4>
        <p className="leading-relaxed dark:text-neutral-400">
          Nostr Wallet Connect uses the existing Nostr infrastructure to allow
          applications to communicate with lightning wallets. This makes it easy
          for developers to integrate and gives the users full flexibility on
          choice.
          <br />
          To learn more about the details have a look at&nbsp;
          <a
            href="https://github.com/getAlby/nips/blob/7-wallet-connect-patch/47.md"
            className="dark:text-gray-300 underline"
          >
            the specs (NIP47)
          </a>
        </p>

        <h4 className="font-bold text-l font-headline mt-6 dark:text-white">
          Which apps support it?
        </h4>
        <p className="leading-relaxed dark:text-neutral-400">
          <a
            href="https://linktr.ee/amethyst.social"
            className="dark:text-gray-300 underline"
          >
            Amethyst
          </a>
          &nbsp;is the first Nostr Client to fully support Nostr Wallet Connect.
          More are working on support for it and are following shortly.
        </p>

        <h4 className="font-bold text-l font-headline mt-6 dark:text-white">
          Is it open source?
        </h4>
        <p className="leading-relaxed dark:text-neutral-400">
          Of course! Nostr Wallet Connect is an open standard and our
          implementation is&nbsp;
          <a
            href="https://github.com/getAlby/nostr-wallet-connect/"
            className="dark:text-gray-300 underline"
          >
            open source on GitHub
          </a>
        </p>

        <h4 className="font-bold text-l font-headline mt-6 dark:text-white">
          Can I run my own?
        </h4>
        <p className="leading-relaxed dark:text-neutral-400">
          Yes, Nostr Wallet Connect is for every wallet and can easily be
          self-hosted.
          <br />
          We are looking for contributors to add support for more wallet
          backends. If you're interested please contact us&nbsp;
          <a
            href="https://github.com/getAlby/nostr-wallet-connect/"
            className="dark:text-gray-300 underline"
          >
            on GitHub
          </a>
          &nbsp;or&nbsp;
          <a
            href="mailto:hello@getalby.com"
            className="dark:text-gray-300 underline"
          >
            hello@getalby.com
          </a>
        </p>
      </div>

      <div className="bg-white rounded-md shadow p-4 lg:p-8 dark:bg-surface-02dp">
        <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">
          For developers
        </h3>
        <p className="mb-1 leading-relaxed dark:text-neutral-400">
          Nostr Wallet Connect is an open protocol enabling applications to
          interact with bitcoin lightning wallets. It allows users to connect
          their existing wallets to your application allowing you to easily
          integrate bitcoin lightning functionality but still focus on the core
          functionality of your app.
        </p>

        <ul className="list-inside list-disc mb-8 dark:text-neutral-400">
          <li>
            <a
              href="https://github.com/getAlby/nips/blob/master/47.md"
              className="dark:text-gray-300 underline"
            >
              Read the specs (NIP47)
            </a>
          </li>
          <li>
            <a
              href="https://github.com/getAlby/alby-js-sdk#nostr-wallet-connect-documentation"
              className="dark:text-gray-300 underline"
            >
              Look at the JavaScript SDK
            </a>
          </li>
        </ul>

        <p className="sm:text-2xl font-bold font-headline mb-4 text-center dark:text-white">
          Happy zapping!
          <img
            src={nostrBitcoin}
            alt="Bitcoin on Nostr"
            className="w-80 mx-auto"
          />
        </p>
      </div>
    </>
  );
}

export default About;
