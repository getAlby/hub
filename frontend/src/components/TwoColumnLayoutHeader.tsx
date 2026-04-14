type Props = {
  title: string;
  description: string;
  pageTitle?: string;
};

export default function TwoColumnLayoutHeader({
  title,
  description,
  pageTitle,
}: Props) {
  return (
    <>
      {pageTitle && <title>{`${pageTitle} · Alby Hub`}</title>}
      <div className="grid gap-2 text-center">
        <h1 className="text-2xl font-semibold">{title}</h1>
        <p className="text-muted-foreground">{description}</p>
      </div>
    </>
  );
}
