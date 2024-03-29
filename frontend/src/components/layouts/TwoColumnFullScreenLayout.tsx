import { Outlet } from "react-router-dom";
import { ModeToggle } from "src/components/ui/mode-toggle";

export default function TwoColumnFullScreenLayout() {
  const quotes = [
    {
      content: `This isn't about nation-states anymore. This isn't about who adopts
          bitcoin first or who adopts cryptocurrencies first, because the
          internet is adopting cryptocurrencies, and the internet is the world's
          largest economy. It is the first transnational economy, and it needs a
          transnational currency.`,
      author: "Andreas M. Antonopoulos",
    },
    {
      content: `It might make sense just to get some in case it catches on. If enough people think the same way, that becomes a self fulfilling prophecy. Once it gets bootstrapped, there are so many applications if you could effortlessly pay a few cents to a website as easily as dropping coins in a vending machine.`,
      author: "Satoshi Nakamoto",
    },
    {
      content: `Since we’re all rich with bitcoins, or we will be once they’re worth a million dollars like everyone expects, we ought to put some of this unearned wealth to good use.`,
      author: "Hal Finney",
    },
    {
      content: `I don't believe we shall ever have a good money again before we take the thing out of the hands of government, that is, we can't take it violently out of the hands of government, all we can do is by some sly roundabout way introduce something that they can't stop.`,
      author: "Friedrich August von Hayek",
    },
  ];

  const quote = quotes[Math.floor(Math.random() * quotes.length)];

  return (
    <div className="w-full lg:grid lg:min-h-screen lg:grid-cols-2 items-stretch text-background">
      <div className="hidden lg:flex flex-col bg-foreground  justify-end p-10 gap-2">
        <div className="flex-1 w-full">
          <div className="flex flex-row justify-between">
            <h1 className="text-lg font-medium">Alby Hub</h1>
            <ModeToggle />
          </div>
        </div>
        <p className="text-muted-foreground text-lg">{quote.content}</p>
        <p className="text-background text-sm">{quote.author}</p>
      </div>
      <div className="flex items-center justify-center py-12 text-foreground">
        <Outlet />
      </div>
    </div>
  );
}
