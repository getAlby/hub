export default function NewOnchainAddress() {
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <p>You can deposit Bitcoin to the below address:</p>
      <input
        className="w-full font-mono shadow-md"
        value="pubkey@domain_name:9735"
      ></input>
    </div>
  );
}
