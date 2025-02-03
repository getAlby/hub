import { Code, PlusCircle, RefreshCw } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

function SupportAlby() {
  return (
    <>
      <div className="min-h-screen max-w-screen-sm mx-auto flex flex-col justify-center">
        <div className="flex flex-col items-center justify-center gap-6">
          <section style={{ textAlign: "center" }}>
            <h2 className="text-3xl font-semibold mb-2">
              ✨ Your Support Matters
            </h2>
            <p className="text-muted-foreground">
              Our open-source Lightning node is dedicated to enhancing the
              Bitcoin ecosystem by providing a reliable, efficient, and
              user-friendly platform for transactions. With your support, we can
              continue to innovate and expand our services.
            </p>
          </section>

          <Card className="w-full">
            <CardHeader>
              <CardTitle>Why Your Contribution Is Important</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="flex flex-col gap-2">
                <li className="flex flex-col ">
                  <div className="flex flex-row items-center">
                    <PlusCircle className="w-4 h-4 mr-2" />
                    Enhanced Features
                  </div>
                  <div className="text-muted-foreground">
                    Your support allows us to build and integrate new features
                  </div>
                </li>
                <li className="flex flex-col ">
                  <div className="flex flex-row items-center">
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Regular Updates
                  </div>
                  <div className="text-muted-foreground">
                    Contributions fund ongoing maintenance and improvements
                  </div>
                </li>
                <li className="flex flex-col ">
                  <div className="flex flex-row items-center">
                    <Code className="w-4 h-4 mr-2" />
                    Open-Source
                  </div>
                  <div className="text-muted-foreground">
                    Keep Alby Hub open-source and free for all users
                  </div>
                </li>
              </ul>
            </CardContent>
          </Card>

          <Card className="w-full">
            <CardHeader>
              <CardTitle>Join Our Community of Supporters</CardTitle>
            </CardHeader>
            <CardContent>
              <p>
                When you set up a recurring donation, you join a dedicated group
                of enthusiasts who value the ongoing success and development of
                Alby Hub. Together, we can make a difference!
              </p>

              <div className="flex flex-row gap-2 mt-5">
                {Array.from({ length: 12 }).map((_, index) => (
                  <Avatar key={index}>
                    <AvatarImage
                      src="https://via.placeholder.com/100"
                      alt={`Supporter ${index + 1}`}
                    />
                    <AvatarFallback>{`S${index + 1}`}</AvatarFallback>
                  </Avatar>
                ))}
              </div>
            </CardContent>
          </Card>

          <Button size="lg">Become a Supporter</Button>
        </div>
      </div>
    </>
  );
}

export default SupportAlby;
