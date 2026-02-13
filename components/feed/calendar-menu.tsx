"use client";

import { Calendar, ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useRef, useState } from "react";

interface CalendarMenuProps {
	currentDate?: string;
	navigateToDate?: (date: string) => void;
	availableDates?: string[];
}

export function CalendarMenu({
	currentDate,
	navigateToDate,
	availableDates = [],
}: CalendarMenuProps) {
	const [isOpen, setIsOpen] = useState(false);
	const [displayMonth, setDisplayMonth] = useState<Date>(() => {
		if (currentDate) {
			return new Date(currentDate);
		}
		return new Date();
	});
	const containerRef = useRef<HTMLDivElement>(null);

	const daysInMonth = (date: Date) => {
		return new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate();
	};

	const firstDayOfMonth = (date: Date) => {
		return new Date(date.getFullYear(), date.getMonth(), 1).getDay();
	};

	const formatDateForAPI = (day: number) => {
		const year = displayMonth.getFullYear();
		const month = String(displayMonth.getMonth() + 1).padStart(2, "0");
		const date = String(day).padStart(2, "0");
		return `${year}-${month}-${date}`;
	};

	const handlePrevMonth = () => {
		setDisplayMonth(
			new Date(displayMonth.getFullYear(), displayMonth.getMonth() - 1),
		);
	};

	const handleNextMonth = () => {
		setDisplayMonth(
			new Date(displayMonth.getFullYear(), displayMonth.getMonth() + 1),
		);
	};

	const isDateAvailable = (day: number) => {
		const dateStr = formatDateForAPI(day);
		// Only show dates as available if they're in the availableDates array
		// If availableDates is empty (fetch failed or not loaded), grey out all dates
		return availableDates.length > 0 && availableDates.includes(dateStr);
	};

	const isToday = (day: number) => {
		const date = new Date(
			displayMonth.getFullYear(),
			displayMonth.getMonth(),
			day,
		);
		const today = new Date();
		return (
			date.getDate() === today.getDate() &&
			date.getMonth() === today.getMonth() &&
			date.getFullYear() === today.getFullYear()
		);
	};

	const isCurrentDate = (day: number) => {
		if (!currentDate) return false;
		const date = new Date(currentDate);
		return (
			day === date.getDate() &&
			displayMonth.getMonth() === date.getMonth() &&
			displayMonth.getFullYear() === date.getFullYear()
		);
	};

	const days = [];
	const totalDays = daysInMonth(displayMonth);
	const firstDay = firstDayOfMonth(displayMonth);

	// Empty cells for days before month starts
	for (let i = 0; i < firstDay; i++) {
		days.push(null);
	}

	// Days of the month
	for (let i = 1; i <= totalDays; i++) {
		days.push(i);
	}

	const monthName = displayMonth.toLocaleString("en-US", { month: "short" });
	const year = displayMonth.getFullYear();

	const handleDateSelect = (dateStr: string) => {
		navigateToDate?.(dateStr);
		setIsOpen(false);
	};

	// Close on outside click
	useEffect(() => {
		const handleClickOutside = (event: MouseEvent) => {
			if (
				containerRef.current &&
				!containerRef.current.contains(event.target as Node)
			) {
				setIsOpen(false);
			}
		};

		if (isOpen) {
			document.addEventListener("mousedown", handleClickOutside);
			return () =>
				document.removeEventListener("mousedown", handleClickOutside);
		}
	}, [isOpen]);

	return (
		<div ref={containerRef} className="flex items-center">
			<button
				type="button"
				onClick={() => setIsOpen(!isOpen)}
				className="icon-btn cursor-pointer"
				aria-label="Open calendar"
				aria-expanded={isOpen}
			>
				<Calendar className="h-5 w-5 icon-color" />
			</button>

			{isOpen && (
				<div className="absolute top-full right-0 mt-2 bg-card/80 backdrop-blur-md rounded-lg shadow-lg p-3 z-50 w-72 border border-border/50">
					<div className="flex items-center justify-between mb-3">
						<button
							type="button"
							onClick={handlePrevMonth}
							className="icon-btn-sm cursor-pointer"
							aria-label="Previous month"
						>
							<ChevronLeft className="w-4 h-4 icon-color" />
						</button>
						<h3 className="font-semibold text-sm text-foreground">
							{monthName} {year}
						</h3>
						<button
							type="button"
							onClick={handleNextMonth}
							className="icon-btn-sm cursor-pointer"
							aria-label="Next month"
						>
							<ChevronRight className="w-4 h-4 icon-color" />
						</button>
					</div>

					<div className="grid grid-cols-7 gap-0.5 mb-2">
						{["S", "M", "T", "W", "T", "F", "S"].map((day, index) => (
							<div
								key={index}
								className="text-center text-xs font-semibold text-muted-foreground py-1"
							>
								{day}
							</div>
						))}
					</div>

					<div className="grid grid-cols-7 gap-0.5">
						{days.map((day) =>
							day === null ? (
								<div key={`empty-${Math.random()}`} />
							) : (
								<button
									type="button"
									key={day}
									onClick={() => {
										handleDateSelect(formatDateForAPI(day));
									}}
									disabled={!isDateAvailable(day)}
									className={`
										py-1 px-0.5 rounded text-xs font-medium transition-colors
										${
											isCurrentDate(day)
												? "bg-accent text-accent-foreground cursor-pointer"
												: isToday(day)
													? "border border-accent text-foreground cursor-pointer"
													: !isDateAvailable(day)
														? "text-muted-foreground/50 cursor-not-allowed"
														: "hover:bg-accent text-foreground cursor-pointer"
										}
									`}
								>
									{day}
								</button>
							),
						)}
					</div>
				</div>
			)}
		</div>
	);
}
