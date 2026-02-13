"use client";

import type React from "react";
import { Section } from "./section";

type NavbarSectionProps = {
	children: React.ReactNode;
};

const NAVBAR_SECTION_CLASSES = "py-4 my-0 gap-0";

export function NavbarSection({ children }: NavbarSectionProps) {
	return <Section className={NAVBAR_SECTION_CLASSES}>{children}</Section>;
}
