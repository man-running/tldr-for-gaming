"use client";

import { usePathname } from "next/navigation";
import { Navbar } from "./navbar";
import { NavbarSection } from "./navbar-section";

export function RootNavbar() {
	const pathname = usePathname();
	const isHomePage = pathname === "/";
	const isGamingPage = pathname.startsWith("/gaming");

	// Home and gaming pages render their own navbar inside their context providers
	if (isHomePage || isGamingPage) {
		return null;
	}

	return (
		<NavbarSection>
			<Navbar />
		</NavbarSection>
	);
}
