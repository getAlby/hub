export default function Container({ children }: React.PropsWithChildren) {
  return (
    <div className="flex flex-col items-center w-full p-2 mt-6">
      <div className="max-w-md w-full flex flex-col items-center">
        {children}
      </div>
    </div>
  );
}
