import { DigestNavbarProvider } from "@/components/feed/digest-navbar-provider";
import { Navbar } from "@/components/layout/navbar";
import { NavbarSection } from "@/components/layout/navbar-section";

export default function GamingLayout({ children }: { children: React.ReactNode }) {
  return (
    <DigestNavbarProvider>
      <NavbarSection>
        <Navbar />
      </NavbarSection>
      {children}
    </DigestNavbarProvider>
  );
}
