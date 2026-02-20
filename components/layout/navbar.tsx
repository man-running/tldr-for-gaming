"use client";

import { NavbarLayout } from "@takara/shared-ui/navbar-layout";
import { ArrowLeft, Search } from "lucide-react";
import dynamic from "next/dynamic";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
	type ReactNode,
	useCallback,
	useContext,
	useEffect,
	useRef,
} from "react";
import { PaperSearchPanelMemo } from "@/components/search/paper-search-panel";
import { useDebounce } from "@/lib/hooks/use-debounce";
import { usePaperSearch } from "@/lib/hooks/use-paper-search";
import { useSearchStore } from "@/lib/stores/search-store";
import { cn } from "@/lib/utils";
import { FavoritesMenu } from "./favorites-menu";
import { NavbarFeedContext } from "./home-page-navbar-provider";

const ThemeToggle = dynamic(
	() => import("@/components/ui/theme-toggle").then((m) => m.ThemeToggle),
	{ ssr: false },
);

const CalendarMenu = dynamic(
	() => import("@/components/feed/calendar-menu").then((m) => m.CalendarMenu),
	{ ssr: false },
);

interface NavbarFeedData {
	currentDate?: string;
	navigateToDate?: (date: string) => void;
	availableDates?: string[];
}

type SearchStage = "idle" | "input" | "results";

const SEARCH_EVENT_ORDER: Record<
	SearchStage,
	{
		iconTransform: string;
		inputVisible: boolean;
		panelVisible: boolean;
		description: string;
	}
> = {
	idle: {
		iconTransform: "",
		inputVisible: false,
		panelVisible: false,
		description: "Icon at far right, search input collapsed, panel hidden.",
	},
	input: {
		iconTransform: "-translate-x-1",
		inputVisible: true,
		panelVisible: false,
		description: "Icon slides left, input reveals, waiting for user query.",
	},
	results: {
		iconTransform: "-translate-x-2",
		inputVisible: true,
		panelVisible: true,
		description:
			"Icon locked left, input stays open, panel emphasizes results.",
	},
};

function Logo({ hideOnMobile }: { hideOnMobile?: boolean }) {
	return (
		<Link
			href="/"
			className={cn(
				"flex items-center text-[25px] leading-[191%] tracking-[0.025em]",
				hideOnMobile && "hidden sm:flex",
			)}
		>
			<span className="font-black text-foreground">tldr.</span>
			<span className="font-black text-accent">takara.ai</span>
		</Link>
	);
}

interface SearchBarProps {
	stage: SearchStage;
	query: string;
	onOpen: () => void;
	onChange: (value: string) => void;
	inputRef: React.RefObject<HTMLInputElement | null>;
}

function SearchBar({
	stage,
	query,
	onOpen,
	onChange,
	inputRef,
}: SearchBarProps) {
	const stageConfig = SEARCH_EVENT_ORDER[stage];
	return (
		<div className="flex items-center justify-end">
			<button
				type="button"
				onClick={onOpen}
				aria-label="Search papers"
				className={cn(
					"icon-btn transition-transform duration-300 cursor-pointer",
					stageConfig.iconTransform,
				)}
			>
				<Search className="h-5 w-5 icon-color" />
			</button>
			<div
				className={cn(
					"transition-all duration-300 ease-out overflow-hidden min-w-0",
					stageConfig.inputVisible
						? "max-w-[220px] sm:max-w-[320px] opacity-100"
						: "max-w-0 opacity-0 pointer-events-none",
				)}
			>
				<input
					ref={inputRef}
					type="text"
					value={query}
					onChange={(event) => onChange(event.target.value)}
					placeholder="Search papers"
					className="w-full border-b border-border bg-transparent text-base text-primary-label placeholder:text-secondary-label py-1 focus:border-accent outline-none"
				/>
			</div>
		</div>
	);
}

function HomePageNav({
	feedData,
	searchSlot,
	searchStage,
}: {
	feedData: NavbarFeedData;
	searchSlot: ReactNode;
	searchStage: SearchStage;
}) {
	return (
		<header className="w-full h-[75px] sm:h-[93px]">
			<NavbarLayout className="flex items-center justify-between gap-4 flex-wrap">
				<Logo hideOnMobile={searchStage !== "idle"} />
				<div className="flex items-center gap-2 sm:gap-4">
					{searchSlot}
					<FavoritesMenu />
					<CalendarMenu
						currentDate={feedData.currentDate}
						navigateToDate={feedData.navigateToDate}
						availableDates={feedData.availableDates}
					/>
					<ThemeToggle />
				</div>
			</NavbarLayout>
		</header>
	);
}

function PaperPageNav({
	router,
	searchSlot,
	searchStage,
}: {
	router: ReturnType<typeof useRouter>;
	searchSlot: ReactNode;
	searchStage: SearchStage;
}) {
	const handleBack = useCallback(() => {
		if (typeof window === "undefined") return;

		const referrer = document.referrer;
		const currentOrigin = window.location.origin;
		const hasInternalReferrer =
			referrer &&
			referrer.startsWith(currentOrigin) &&
			referrer !== window.location.href;

		if (hasInternalReferrer) {
			router.back();
		} else {
			router.push("/");
		}
	}, [router]);

	return (
		<header className="w-full h-[75px] sm:h-[93px]">
			<NavbarLayout className="flex items-center justify-between gap-4 flex-wrap">
				<button
					type="button"
					onClick={handleBack}
					aria-label="Go back to previous page"
					className="icon-btn cursor-pointer"
				>
					<ArrowLeft className="h-5 w-5 icon-color" />
				</button>
				<Logo hideOnMobile={searchStage !== "idle"} />
				<div className="flex items-center gap-2 sm:gap-4">
					{searchSlot}
					<FavoritesMenu />
					<ThemeToggle />
				</div>
			</NavbarLayout>
		</header>
	);
}

function DefaultNav({
	searchSlot,
	searchStage,
}: {
	searchSlot: ReactNode;
	searchStage: SearchStage;
}) {
	return (
		<header className="w-full h-[75px] sm:h-[93px]">
			<NavbarLayout className="flex items-center justify-between gap-4 flex-wrap">
				<Logo hideOnMobile={searchStage !== "idle"} />
				<div className="flex items-center gap-2 sm:gap-4">
					{searchSlot}
					<FavoritesMenu />
					<ThemeToggle />
				</div>
			</NavbarLayout>
		</header>
	);
}

export function Navbar() {
	const pathname = usePathname();
	const router = useRouter();
	const feedData = useContext(NavbarFeedContext);

	const { searchStage, query, setQuery, closeSearch, openSearch } =
		useSearchStore();

	const inputRef = useRef<HTMLInputElement>(null);

	const trimmedQuery = query.trim();
	const hasQuery = query.length > 0;
	const debouncedQuery = useDebounce(trimmedQuery, 400);

	const { performSearch, loadQueryEmbedding, abort } = usePaperSearch({
		debouncedQuery,
		searchStage,
		hasQuery,
	});

	const isHomePage = pathname === "/";
	const isGamingPage = pathname.startsWith("/gaming");
	const isPaperPage = pathname.startsWith("/p/");

	useEffect(() => {
		if (searchStage !== "idle" && inputRef.current) {
			inputRef.current.focus();
		}
	}, [searchStage]);

	const handleCloseSearch = useCallback(() => {
		closeSearch();
		abort();
	}, [closeSearch, abort]);

	useEffect(() => {
		if (searchStage === "idle") return;

		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") handleCloseSearch();
		};

		document.addEventListener("keydown", handleKeyDown);
		return () => document.removeEventListener("keydown", handleKeyDown);
	}, [searchStage, handleCloseSearch]);

	useEffect(() => {
		loadQueryEmbedding();
	}, [loadQueryEmbedding]);

	useEffect(() => {
		performSearch();
	}, [performSearch]);

	// Search is temporarily hidden — to be re-implemented with DS1
	// const searchSlot = (
	// 	<SearchBar
	// 		stage={searchStage}
	// 		query={query}
	// 		onOpen={openSearch}
	// 		onChange={setQuery}
	// 		inputRef={inputRef}
	// 	/>
	// );
	const searchSlot = null;

	return (
		<>
			{(isHomePage || isGamingPage) && feedData ? (
				<HomePageNav
					feedData={feedData}
					searchSlot={searchSlot}
					searchStage={searchStage}
				/>
			) : isPaperPage ? (
				<PaperPageNav
					router={router}
					searchSlot={searchSlot}
					searchStage={searchStage}
				/>
			) : (
				<DefaultNav searchSlot={searchSlot} searchStage={searchStage} />
			)}

			<PaperSearchPanelMemo onNavigate={handleCloseSearch} />
		</>
	);
}
