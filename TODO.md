# TODO List

## UI/Branding Improvements

### 1. Update Digest Banner Colors
**Priority:** Medium
**Description:** Change the default banner to use colors more in line with iGaming/TLDR branding
**Current:** Default theme colors
**Target:** Brand-aligned color scheme
**Files to modify:**
- `components/feed/digest-banner.tsx` (or similar)
- Possibly `app/digest/page.tsx`

---

### 2. Style "Read Full Article" Button
**Priority:** Medium
**Description:** Update the "Read Full Article" button to use Takara red branding
**Current:** Default button styling
**Target:** Takara red (#FF0000 or brand-specific red)
**Files to modify:**
- `components/feed/ranked-article-card.tsx`
- Check for any other article card components

---

### 3. Hide Search Feature
**Priority:** High
**Description:** Temporarily hide the search functionality and re-implement later using DS1 (Design System 1)
**Current:** Search is visible/enabled
**Target:** Hidden from UI, code preserved for future re-implementation
**Files to modify:**
- `components/layout/root-navbar.tsx` (or navigation component)
- Consider adding feature flag: `FEATURE_SEARCH_ENABLED=false`

**Notes:**
- Don't delete search code, just hide/disable it
- Plan to rebuild with DS1 design system
- May need to update navigation/header layout

---

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

## Future Enhancements (Not in this TODO)
- Analytics/tracking
- User accounts
- Personalized digest preferences
- Multiple digest categories (casino, sports betting, regulation, etc.)
- RSS feed generation
- Social media sharing
