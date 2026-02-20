"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { NavbarFeedContext } from "@/components/layout/home-page-navbar-provider";

export function DigestNavbarProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [availableDates, setAvailableDates] = useState<string[]>([]);

  // Extract current date from /gaming/[date] or default to today
  const dateMatch = pathname.match(/^\/gaming\/(\d{4}-\d{2}-\d{2})$/);
  const currentDate = dateMatch ? dateMatch[1] : new Date().toISOString().split("T")[0];

  useEffect(() => {
    fetch("/api/digest/dates")
      .then((res) => res.json())
      .then((data) => setAvailableDates(data.dates ?? []))
      .catch(() => setAvailableDates([]));
  }, []);

  const navigateToDate = useCallback((date: string) => {
    window.location.href = `/gaming/${date}`;
  }, []);

  return (
    <NavbarFeedContext.Provider value={{ currentDate, navigateToDate, availableDates }}>
      {children}
    </NavbarFeedContext.Provider>
  );
}
