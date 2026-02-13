"use client";

import { Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

export function ThemeToggle() {
	const [mounted, setMounted] = useState(false);
	const { theme, setTheme, systemTheme } = useTheme();

	// Avoid hydration mismatch by only rendering after mount
	useEffect(() => {
		setMounted(true);
	}, []);

	if (!mounted) {
		return (
			<button
				className="flex items-center justify-center transition-colors"
				aria-label="Toggle theme"
				disabled
				type="button"
			>
				<div className="h-5 w-5" />
			</button>
		);
	}

	const handleThemeToggle = () => {
		if (theme === "system") {
			// If in system mode, go to opposite of system theme
			setTheme(systemTheme === "dark" ? "light" : "dark");
		} else if (theme === "light") {
			// If in light mode, go to dark
			setTheme("dark");
		} else if (theme === "dark") {
			// If in dark mode, return to system
			setTheme("system");
		}
	};

	const getThemeIcon = () => {
		if (theme === "system") {
			return <Monitor className="h-5 w-5 icon-color" aria-hidden="true" />;
		}
		return theme === "dark" ? (
			<Moon className="h-5 w-5 icon-color" aria-hidden="true" />
		) : (
			<Sun className="h-5 w-5 text-amber-500" aria-hidden="true" />
		);
	};

	const getThemeLabel = () => {
		if (theme === "system") {
			return `Switch to ${systemTheme === "dark" ? "light" : "dark"} mode`;
		} else if (theme === "light") {
			return "Switch to dark mode";
		} else {
			return "Switch to system preference";
		}
	};

	return (
		<button
			onClick={handleThemeToggle}
			className="icon-btn cursor-pointer"
			aria-label={getThemeLabel()}
			type="button"
			data-ph-capture-attribute-theme={theme}
			data-ph-capture-attribute-action="toggle-theme"
		>
			{getThemeIcon()}
		</button>
	);
}
