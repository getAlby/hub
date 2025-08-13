function OnchainAddressDisplay({ address }: { address: string }) {
  return (
    <>
      {address.match(/.{1,4}/g)?.map((word, index) => {
        if (index % 2 === 0) {
          return (
            <span key={index} className="text-foreground">
              {word}
            </span>
          );
        } else {
          return (
            <span key={index} className="text-muted-foreground">
              {word}
            </span>
          );
        }
      })}
    </>
  );
}

export default OnchainAddressDisplay;
