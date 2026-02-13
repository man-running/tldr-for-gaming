"use client";

import { usePathname } from "next/navigation";
import { Navbar } from "./navbar";
import { NavbarSection } from "./navbar-section";

export function RootNavbar() {
	const pathname = usePathname();
	const isHomePage = pathname === "/";

	// Only render navbar on non-home pages
	// Home page renders its own navbar inside FeedDataProvider
	if (isHomePage) {
		return null;
	}

	return (
		<NavbarSection>
			<Navbar />
		</NavbarSection>
	);
}
