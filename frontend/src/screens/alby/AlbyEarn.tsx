import { ExternalLinkIcon, HeartIcon, TrophyIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Card, CardContent } from "src/components/ui/card";

import albyExtension from "src/assets/suggested-apps/alby-extension.png";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import alby from "src/assets/suggested-apps/alby.png";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";

interface Platform {
  name: string;
  url: string;
}

interface EarnOpportunity {
  title: string;
  logo: string;
  reward?: number;
  rewardText?: string;
  platforms: Platform[];
}

const earnOpportunities: EarnOpportunity[] = [
  {
    title: "Alby Referral Program",
    logo: alby,
    rewardText: "10% of subscription revenue",
    platforms: [
      {
        name: "getalby.com/referrals",
        url: "https://getalby.com/referrals",
      },
    ],
  },
  {
    title: "Alby Go",
    logo: albyGo,
    reward: 1000,
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
    title: "Alby Extension",
    logo: albyExtension,
    reward: 1000,
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
    title: "Alby",
    logo: alby,
    rewardText: "Our gratitude",
    platforms: [
      {
        name: "Trustpilot",
        url: "https://www.trustpilot.com/review/getalby.com",
      },
    ],
  },
];

export function AlbyEarn() {
  return (
    <>
      <AppHeader
        title="Earn"
        description="Earn bitcoin by referring new users or leaving a review."
        pageTitle="Earn"
      />
      <div className="space-y-6">
        <Alert>
          <TrophyIcon className="h-4 w-4" />
          <AlertTitle>Claim your reward</AlertTitle>
          <AlertDescription className="inline">
            Review one of our products and email your review link or screenshot
            to{" "}
            <ExternalLink
              to="mailto:support@getalby.com"
              className="text-primary hover:underline font-medium"
            >
              support@getalby.com
            </ExternalLink>{" "}
            to claim your reward.
          </AlertDescription>
        </Alert>

        <Card>
          <CardContent>
            <div className="space-y-6">
              {earnOpportunities.map((opportunity) => (
                <div
                  key={opportunity.title}
                  className="flex items-center gap-4 pb-6 last:pb-0 border-b last:border-b-0"
                >
                  <img
                    src={opportunity.logo}
                    className="size-8 md:size-10 rounded-lg shrink-0"
                    alt={opportunity.title}
                  />
                  <div className="flex flex-1 flex-col sm:flex-row sm:items-center gap-1">
                    <div className="flex-1">
                      <div className="font-medium">{opportunity.title}</div>
                      <div className="text-sm text-muted-foreground">
                        {opportunity.platforms.map((platform, index) => (
                          <span key={index}>
                            {index > 0 && " • "}
                            <ExternalLink
                              to={platform.url}
                              className="text-primary hover:underline"
                            >
                              {platform.name}
                              <ExternalLinkIcon className="w-3 h-3 ml-1 inline" />
                            </ExternalLink>
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className="sm:text-right font-medium shrink-0">
                      {opportunity.reward !== undefined ? (
                        <FormattedBitcoinAmount
                          amount={opportunity.reward * 1000}
                        />
                      ) : opportunity.rewardText ? (
                        <span className="text-muted-foreground text-sm inline-flex items-start justify-end gap-1">
                          <HeartIcon className="w-4 h-4 mt-0.5" />
                          {opportunity.rewardText}
                        </span>
                      ) : null}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
