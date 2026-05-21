import {
  ExternalLinkIcon,
  HeartIcon,
  LucideIcon,
  TrophyIcon,
} from "lucide-react";
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
  rewardSat?: number;
  rewardText?: string;
  rewardIcon?: LucideIcon;
  platforms: Platform[];
}

const earnOpportunities: EarnOpportunity[] = [
  {
    title: "Alby Referral Program",
    logo: alby,
    rewardText: "10% of subscription revenue",
    rewardIcon: TrophyIcon,
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
    rewardSat: 1000,
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
    rewardSat: 1000,
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
    rewardIcon: HeartIcon,
    platforms: [
      {
        name: "Trustpilot",
        url: "https://www.trustpilot.com/review/getalby.com",
      },
    ],
  },
];

function Reward({ opportunity }: { opportunity: EarnOpportunity }) {
  if (opportunity.rewardSat !== undefined) {
    return (
      <FormattedBitcoinAmount
        amountMsat={opportunity.rewardSat * 1000}
        className="text-lg font-semibold tabular-nums"
      />
    );
  }

  if (opportunity.rewardText && opportunity.rewardIcon) {
    const Icon = opportunity.rewardIcon;
    return (
      <span className="inline-flex items-center gap-1.5 text-sm text-muted-foreground">
        <Icon className="size-4" />
        {opportunity.rewardText}
      </span>
    );
  }

  return null;
}

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
              className="underline font-medium"
            >
              support@getalby.com
            </ExternalLink>{" "}
            to claim your reward.
          </AlertDescription>
        </Alert>

        <Card>
          <CardContent>
            <div className="divide-y">
              {earnOpportunities.map((opportunity) => (
                <div
                  key={opportunity.title}
                  className="flex items-center gap-4 py-5 first:pt-0 last:pb-0"
                >
                  <img
                    src={opportunity.logo}
                    className="size-10 md:size-12 rounded-lg shrink-0"
                    alt={opportunity.title}
                  />
                  <div className="flex flex-1 flex-col sm:flex-row sm:items-center gap-2 min-w-0">
                    <div className="flex-1 min-w-0">
                      <div className="font-medium">{opportunity.title}</div>
                      <div className="text-sm text-muted-foreground flex flex-wrap items-center gap-x-2">
                        {opportunity.platforms.map((platform, index) => (
                          <span
                            key={index}
                            className="inline-flex items-center"
                          >
                            {index > 0 && <span className="mr-2">•</span>}
                            <ExternalLink
                              to={platform.url}
                              className="underline text-foreground inline-flex items-center hover:text-muted-foreground"
                            >
                              {platform.name}
                              <ExternalLinkIcon className="size-3 ml-1" />
                            </ExternalLink>
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className="sm:text-right shrink-0">
                      <Reward opportunity={opportunity} />
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
