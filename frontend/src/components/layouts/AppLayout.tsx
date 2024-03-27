import { ModeToggle } from "src/components/ui/mode-toggle";
import {
  MessageCircle,
  Cable,
  LayoutGrid,
  Wallet,
  Settings,
  CircleHelp,
  ShieldCheck,
} from "lucide-react";

import { Link, Outlet } from "react-router-dom";
import { Avatar, AvatarImage, AvatarFallback } from "src/components/ui/avatar";

export default function AppLayout() {
  return (
    <div className="font-sans grid min-h-screen w-full md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
      <div className="hidden border-r bg-muted/40 md:block">
        <div className="flex h-full max-h-screen flex-col gap-2">
          <div className="flex-1">
            <nav className="grid items-start px-2 py-2 text-sm font-medium lg:px-4">
              <div className="p-3 ">
                <Link to="/" className="font-semibold text-xl">
                  <span className="">Alby Hub</span>
                </Link>
              </div>
              <Link
                to="/wallet"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
              >
                <Wallet className="h-4 w-4" />
                Wallet
              </Link>
              <Link
                to="/connections"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
              >
                <Cable className="h-4 w-4" />
                Connections
              </Link>
              <Link
                to="/apps"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
              >
                <LayoutGrid className="h-4 w-4" />
                Apps
              </Link>
              <Link
                to="#"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary cursor-not-allowed"
              >
                <ShieldCheck className="h-4 w-4" />
                Permissions
              </Link>
            </nav>
          </div>
          <div className="flex flex-col">
            <nav className="grid items-start p-2 text-sm font-medium lg:px-4">
              <ModeToggle />
              <Link
                to="/settings"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
              >
                <Settings className="h-4 w-4" />
                Settings
              </Link>
              <Link
                to="#"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary cursor-not-allowed"
              >
                <CircleHelp className="h-4 w-4" />
                Help
              </Link>
              <Link
                to="#"
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary cursor-not-allowed"
              >
                <MessageCircle className="h-4 w-4" />
                Leave Feedback
              </Link>
            </nav>
            <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6 gap-3 border-t border-border">
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src="https://github.com/shadcn.png"
                  alt="@shadcn"
                />
                <AvatarFallback>CN</AvatarFallback>
              </Avatar>
              <Link
                to="/"
                className="flex items-center gap-2 font-semibold text-lg"
              >
                John Doe
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* <header className="flex h-14 items-center gap-4 border-b bg-muted/40 px-4 lg:h-[60px] lg:px-6">
          <Sheet>
            <SheetTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="shrink-0 md:hidden"
              >
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle navigation menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="flex flex-col">
              <nav className="grid gap-2 text-lg font-medium">
                <Link
                  to="#"
                  className="flex items-center gap-2 text-lg font-semibold"
                >
                  <Package2 className="h-6 w-6" />
                  <span>Alby Hub</span>
                </Link>
                TODO: Mobile menu
              </nav>
            </SheetContent>
          </Sheet>
          <div className="w-full flex-1"></div>
          <ModeToggle />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="secondary" size="icon" className="rounded-full">
                <CircleUser className="h-5 w-5" />
                <span className="sr-only">Toggle user menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>My Account</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <Link to="/channels">Channels</Link>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Link to="/backup/mnemonic">Backups</Link>
              </DropdownMenuItem>
              <DropdownMenuItem>Support</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Logout</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </header> */}
      <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-8">
        <Outlet />
      </main>
    </div>
  );
}
