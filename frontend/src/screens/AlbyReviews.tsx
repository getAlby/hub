import { ExternalLinkIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

import albyGo from "src/assets/suggested-apps/alby-go.png";
import albyExtension from "src/assets/suggested-apps/alby.png";

interface Platform {
  name: string;
  url: string;
}

interface ProductOpportunity {
  id: string;
  title: string;
  logo: string;
  reward: string;
  platforms: Platform[];
}

const productOpportunities: ProductOpportunity[] = [
  {
    id: "alby-go",
    title: "Alby Go",
    logo: albyGo,
    reward: "1,000 sats",
    platforms: [
      {
        name: "Google Play",
        url: "https://play.google.com/store/apps/details?id=com.getalby.mobile",
      },
      {
        name: "App Store",
        url: "https://apps.apple.com/app/alby-go/id6476033149",
      },
    ],
  },
  {
    id: "alby-extension",
    title: "Alby Extension",
    logo: albyExtension,
    reward: "1,000 sats",
    platforms: [
      {
        name: "Chrome",
        url: "https://chrome.google.com/webstore/detail/alby-bitcoin-wallet-for-l/iokeahhehimjnekafflcihljlcjccdbe",
      },
      {
        name: "Firefox",
        url: "https://addons.mozilla.org/en-US/firefox/addon/alby/",
      },
    ],
  },
  {
    id: "alby-trustpilot",
    title: "Alby",
    logo: albyExtension,
    reward: "2,000 sats",
    platforms: [
      {
        name: "Trustpilot",
        url: "https://www.trustpilot.com/review/getalby.com",
      },
    ],
  },
];

export function AlbyReviews() {
  return (
    <>
      <AppHeader title="Earn Bitcoin" />

      <div className="space-y-8">
        <Card>
          <CardHeader>
            <CardTitle>Write a review, earn bitcoin</CardTitle>
            <CardDescription>
              Help others discover Alby by sharing your experience. Send your
              review link to{" "}
              <ExternalLink
                to="mailto:hello@getalby.com"
                className="text-primary"
              >
                hello@getalby.com
              </ExternalLink>{" "}
              to receive sats.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {productOpportunities.map((product) => (
                <div key={product.id} className="flex items-center gap-4">
                  <img
                    src={product.logo}
                    className="w-10 h-10 rounded-lg"
                    alt={product.title}
                  />
                  <div className="flex-1">
                    <div className="font-medium">{product.title}</div>
                    <div className="text-sm text-muted-foreground">
                      {product.platforms.map((platform, index) => (
                        <span key={index}>
                          {index > 0 && " • "}
                          <ExternalLink
                            to={platform.url}
                            className="text-primary"
                          >
                            {platform.name}
                            <ExternalLinkIcon className="w-3 h-3 ml-1 inline" />
                          </ExternalLink>
                        </span>
                      ))}
                    </div>
                  </div>
                  <div className="text-right font-medium">{product.reward}</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
