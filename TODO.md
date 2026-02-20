# TODO List

## Features

### 4. Implement Email Subscriptions
**Priority:** High
**Description:** Research and implement email subscription feature similar to TLDR newsletter
**Current:** No subscription functionality
**Target:** Daily digest email subscriptions

**Research Questions:**
- How does TLDR.tech handle subscriptions?
- What email service to use? (SendGrid, Resend, Mailchimp, ConvertKit?)
- When to send emails? (Trigger after daily digest generation)
- What should email template look like?

**Technical Requirements:**
- Database/storage for subscriber emails
- Email sending service integration
- Subscription form component
- Unsubscribe mechanism (required by law)
- Email template matching digest format

**Implementation Steps:**
1. Research TLDR's subscription flow
2. Choose email service provider
3. Design email template
4. Create subscription form UI
5. Add database schema for subscribers
6. Implement subscription API endpoints:
   - POST `/api/subscribe` - Add new subscriber
   - POST `/api/unsubscribe` - Remove subscriber
   - GET `/api/verify-email?token=...` - Verify email address
7. Create email sending job (trigger after digest generation)
8. Add double opt-in verification
9. Implement unsubscribe link in emails
10. Test thoroughly (spam filters, deliverability)

**Compliance Considerations:**
- GDPR compliance (EU users)
- CAN-SPAM Act (US users)
- Include physical address in emails
- Honor unsubscribe requests immediately
- Provide clear privacy policy

**Resources:**
- [TLDR Newsletter](https://tldr.tech/) - Study their subscription flow
- [Resend Documentation](https://resend.com/docs) - Modern email API
- [SendGrid](https://sendgrid.com/) - Established email service
- [React Email](https://react.email/) - Email templates with React

---

### 5. Analytics & Tracking (PostHog)
**Priority:** High
**Description:** PostHog infrastructure is already partially set up — verify it's working and add meaningful custom event tracking.

**Current State:**
- `posthog-js` is installed as a dependency
- `PostHogProvider` exists at `components/analytics/posthog-provider.tsx` and is wired into `app/layout.tsx`
- `/ingest` proxy is configured in `next.config.mjs` (EU region)
- Auto page view and page leave tracking are enabled
- `NEXT_PUBLIC_POSTHOG_KEY` needs to be set in `.env.local` and Vercel env vars

**Implementation Steps:**
1. Confirm `NEXT_PUBLIC_POSTHOG_KEY` is set in `.env.local` and in Vercel project settings
2. Verify events are appearing in the PostHog dashboard (check EU: https://eu.posthog.com)
3. Add custom event tracking for key user actions:
   - Digest page viewed (with date)
   - Article card "Read Full Article" clicked (with article title/URL)
   - Digest navigation (prev/next day)
   - Email subscription CTA clicked
4. Consider adding a `usePostHog` hook wrapper for easy reuse across components

**Files to modify:**
- `.env.local` — add `NEXT_PUBLIC_POSTHOG_KEY`
- `components/feed/ranked-article-card.tsx` — track article clicks
- `components/feed/digest-navigation.tsx` — track navigation events
- `components/email/daily-email-summary/daily-email-subscription-cta.tsx` — track CTA clicks

---

### 6. Wire Calendar Into Digest Navigation
**Priority:** High
**Description:** `CalendarMenu` is fully built and wired into the home page navbar. It needs to be connected to the digest pages so users can navigate between daily digests by date.

**Current State:**
- `CalendarMenu` (`components/feed/calendar-menu.tsx`) is complete — supports `currentDate`, `navigateToDate`, and `availableDates` props, greys out unavailable dates
- `HomePageNavbarProvider` feeds the calendar from the paper feed context (home page only)
- `/digest/[date]/page.tsx` exists and passes `date` to `DigestDisplay` correctly
- `DigestDisplay` already fetches `/api/digest?date={date}` when a date prop is passed
- The digest pages use a plain `Navbar` with no context provider, so the calendar gets `undefined` and is effectively dead on `/digest` and `/digest/[date]`
- `DigestNavigation` (`components/feed/digest-navigation.tsx`) exists as a simpler fallback (native date input + links) but is not rendered anywhere

**What Needs to Be Done:**
1. **Create `/api/digest/dates` endpoint** — list available digest dates from Vercel Blob storage (reuse blob listing logic from existing digest API)
2. **Create `DigestNavbarProvider`** — client component similar to `HomePageNavbarProvider` that:
   - Fetches available dates from `/api/digest/dates`
   - Tracks `currentDate` from the URL param
   - Provides `navigateToDate` using `router.push('/digest/[date]')`
   - Wraps via `NavbarFeedContext`
3. **Wrap digest pages** — add `DigestNavbarProvider` to both `app/digest/page.tsx` and `app/digest/[date]/page.tsx` so the navbar `CalendarMenu` receives live data
4. **Update `Navbar`** — add a `isDigestPage` check (similar to `isHomePage`) so `HomePageNav` (which renders `CalendarMenu`) is used on digest pages too

**Files to modify:**
- `app/api/digest/dates/route.ts` — new endpoint (list available blob dates)
- `components/feed/digest-navbar-provider.tsx` — new provider component
- `app/digest/page.tsx` — wrap with provider
- `app/digest/[date]/page.tsx` — wrap with provider, pass date to provider
- `components/layout/navbar.tsx` — extend page detection to include digest pages

---

## Future Enhancements (Not in this TODO)
- Analytics/tracking
- User accounts
- Personalized digest preferences
- Multiple digest categories (casino, sports betting, regulation, etc.)
- RSS feed generation
- Social media sharing
