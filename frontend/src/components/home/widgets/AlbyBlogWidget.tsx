import { SquareArrowOutUpRightIcon } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { useAlbyBlog } from "src/hooks/useAlbyBlog";
import { cn } from "src/lib/utils";

const fallbackThemes = [
  "from-emerald-200 via-cyan-200 to-yellow-200",
  "from-orange-200 via-amber-100 to-pink-100",
  "from-slate-200 via-zinc-100 to-lime-100",
  "from-sky-200 via-indigo-100 to-violet-100",
];

function hashString(str: string): number {
  return Math.abs([...str].reduce((sum, ch) => sum + ch.charCodeAt(0), 0));
}

export function AlbyBlogWidget() {
  const { data: post } = useAlbyBlog();

  if (!post) {
    return null;
  }

  const theme = fallbackThemes[hashString(post.id) % fallbackThemes.length];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Alby Blog</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="relative h-[247px] overflow-hidden rounded-xl border">
          {post.imageUrl ? (
            <img
              src={post.imageUrl}
              alt={post.title}
              className="absolute inset-0 size-full object-cover"
            />
          ) : (
            <div className={cn("absolute inset-0 bg-gradient-to-br", theme)}>
              <div className="absolute -left-10 top-6 size-36 rounded-full bg-white/35 blur-3xl" />
              <div className="absolute -right-8 bottom-2 size-40 rounded-full bg-white/20 blur-3xl" />
              <div className="absolute inset-0 bg-white/10" />
            </div>
          )}
        </div>
      </CardContent>
      <CardFooter className="flex flex-col items-start gap-4">
        <div className="space-y-1">
          <CardTitle className="text-xl leading-7">{post.title}</CardTitle>
          <CardDescription className="text-base leading-6">
            {post.description}
          </CardDescription>
        </div>
        <div className="flex w-full justify-end">
          <ExternalLinkButton to={post.url} variant="outline">
            Read on Alby Blog
            <SquareArrowOutUpRightIcon />
          </ExternalLinkButton>
        </div>
      </CardFooter>
    </Card>
  );
}
