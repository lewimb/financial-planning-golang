# Frontend Improvement Notes

> **Audit Date:** 2026-05-20
> **Auditor Role:** Senior PM / UX Auditor / Frontend Reviewer
> **App:** Financial Planning SaaS (React Router v7, TanStack Query, Tailwind v4)
> **Branch:** feat/budgets-integration

---

## Executive Summary

### Current Frontend Quality

The app has a **solid architectural foundation** — React Router v7 with SSR, TanStack Query for server state, Zod validation, and a consistent component library. The routing structure, auth flow, and server/client data split are well-designed. The MVP feature set (Transactions, Budgets, Goals, Dashboard, AI Coach) is functionally complete.

However, the app is **not yet production-ready**. Multiple critical flows rely on hard-coded mock data, entire sections lack loading/error/empty states, accessibility is inconsistent, and several destructive actions have no confirmation step. The gap between "it works in development" and "it works reliably for real users" is significant.

### Main Strengths

- Clean route architecture with clear server/client data split
- Strong form accessibility in registration and goal update forms (aria-invalid, aria-describedby)
- Forecast page is an excellent example of comprehensive state handling
- Onboarding wizard is thoughtfully structured
- Consistent use of TanStack Query for cache management
- shadcn/ui + Radix UI provides a solid accessible primitive foundation

### Main Weaknesses

- **Hard-coded mock data** in 5+ production components (Dashboard graph, Budget overview, Financial Health, Key Insights, Goals overview subtitle)
- **No loading skeletons** on most pages — raw null/empty renders on slow connections
- **No error boundaries** — any API failure silently breaks the UI
- **Accessibility gaps** on interactive elements (hover-only delete buttons, unlabeled icon buttons, hardcoded sidebar user info)
- **Missing confirmation dialogs** on destructive actions in Transactions
- **No chat history persistence** — AI Coach resets on every page reload
- **5 static data points** in AI Coach sidebar panels showing fabricated health scores

### Biggest UX Risks

1. Users making financial decisions based on static/fake dashboard data
2. Accidental transaction deletion with no undo or confirmation
3. AI Coach losing conversation context on navigation (breaks the assistant workflow)
4. Onboarding data lost on mid-wizard navigation (user has to restart)
5. Budget/Goal pages silently empty when API fails — looks like "no data" not "error"

### Biggest Business Risks

1. **Trust erosion** — static financial health score (78) and hardcoded insights damage credibility when users notice
2. **Retention failure** — no onboarding guidance means users who skip or fail onboarding hit an empty dashboard with no nudge to populate data
3. **Support cost** — silent API failures leave users confused, triggering support escalations
4. **Regulatory exposure** — no audit log for financial mutations (delete transaction, delete goal, etc.)

---

# 1. Business Improvements

---

## Dashboard — Static Mock Data Displayed as Real

### Problem
`DashboardGraph` and `DashboardBudgetOverview` render hard-coded arrays for the income/expense chart and budget category breakdown. A user with real transactions sees fabricated bar chart data alongside their actual net savings figure.

### Business Impact
- **Trust:** This is a financial app. Users who notice fabricated data immediately question the entire product's accuracy. This is a critical trust issue, not a minor UX bug.
- **Retention:** Users won't return to check a dashboard that doesn't reflect their real activity.
- **Support cost:** Confused users who see chart numbers that don't match their transactions will raise support tickets or churn silently.

### Suggested Improvement
1. Replace `DashboardGraph` hardcoded data with the monthly income/expense from the `/auth/v1/dashboard` loader response — extend the response or add a `/auth/v1/transactions/monthly-summary` endpoint.
2. Replace `DashboardBudgetOverview` with real budget usage from `budget_summary` already available in the dashboard loader.
3. In the interim, hide the graph component entirely if real data isn't available rather than showing fake data.

### Priority
**Critical — fix before any user-facing release**

---

## AI Coach — Fabricated Financial Health Metrics

### Problem
`FinancialHealth` shows a hard-coded score of 78/100 ("Good"), with static sub-metrics for savings rate, budget adherence, and goal progress. `FinancialKeyInsights` displays pre-written insight cards. None of this is derived from the user's actual data.

### Business Impact
- **Trust:** Users comparing their AI Coach score against their actual budget usage will notice the mismatch and lose confidence in the product.
- **Engagement:** If insights never change, users stop returning to the AI Coach tab — it becomes a dead feature.
- **Upsell opportunity lost:** Real personalized insights are a strong upsell vector for premium features.

### Suggested Improvement
1. Wire `FinancialHealth` score to a backend-computed endpoint or derive it client-side from the dashboard loader data (savings rate, budget SAFE/WARNING/EXCEEDED ratio, goal progress percentage).
2. Replace `FinancialKeyInsights` static cards with the `/auth/v1/ml/insights` response already used in the Reports tab.
3. Add a "Last updated" timestamp to health metrics.

### Priority
**High**

---

## Onboarding — No Completion Tracking / Re-Entry Handling

### Problem
The onboarding wizard at `/onboarding` has no mechanism to resume mid-completion. If a user drops off after Step 1, they restart from Step 1 with empty fields. If they skip onboarding (the check is a redirect on 404), there is no prompt inside the app to complete it later.

### Business Impact
- **Activation:** Users who don't complete onboarding miss the financial profile setup, reducing the quality of AI insights and forecast accuracy.
- **Conversion:** Low activation rates directly reduce feature engagement metrics.
- **Retention:** Users who hit an empty dashboard with no nudge to add data have high early churn.

### Suggested Improvement
1. Persist Step 1 values in `sessionStorage` so back-navigation doesn't clear them.
2. Add an onboarding progress banner in the Dashboard that shows when financial profile is incomplete (currently the app redirects to onboarding on 404, but a softer in-app prompt handles partial completion better).
3. Add a "Complete your profile" card in Settings → Financial Profile as a fallback entry point.

### Priority
**High**

---

## Sidebar — Hardcoded User Profile

### Problem
`AppSidebar` footer displays "John Doe" and "johndoe@email.com" hardcoded. Every logged-in user sees this fake profile.

### Business Impact
- **Trust / Professionalism:** A financial app showing the wrong user name in the nav is a serious UX credibility failure.
- **Support cost:** Users think they are logged into someone else's account.

### Suggested Improvement
1. The auth layout loader already validates the JWT and has `payload` with `userId`. Fetch the user profile during the auth layout load and pass user name/email to `AppSidebar` via context or props.
2. Show the user's actual avatar initial (first letter of name) in the `Avatar` component.

### Priority
**Critical**

---

## Goals — "2 Completed This Year" Hardcoded Subtitle

### Problem
`GoalsOverview` renders a static subtitle "2 completed this year" regardless of actual completed goals.

### Business Impact
- **Trust:** Users with 0 or 5 completed goals see wrong numbers.
- **Engagement:** Accurate progress metrics motivate users; fake ones don't.

### Suggested Improvement
Derive the completed-this-year count from the goals overview response (`goal_summary.completed` already exists in the dashboard data). Pass it as a dynamic prop to `GoalsOverview`.

### Priority
**High**

---

## Missing: Empty States with Actionable CTAs

### Problem
When users have no transactions, budgets, or goals, most pages render blank cards or empty grids with no guidance. The only page with a proper empty state is `GoalsList`.

### Business Impact
- **Activation:** New users hitting empty screens have no next action — they churn.
- **Engagement:** Empty states with CTAs directly drive feature adoption.

### Suggested Improvement
Add illustrated empty states with action buttons on:
- Dashboard (no transactions yet → "Add your first transaction")
- Transactions table (empty → "Record your first income or expense")
- Budgets grid (no budgets → "Create a budget for a category")
- Reports (no data → "Add transactions to see your reports")

Each empty state should include a short description of the feature's value, not just "No data found."

### Priority
**High**

---

## Missing: Transaction Edit Functionality

### Problem
`TransactionColumns` has edit functionality commented out. Users can create and delete transactions but cannot edit them. Any data entry error requires deletion and re-creation.

### Business Impact
- **User satisfaction:** Inability to edit is a basic usability gap. Users with categorization errors must delete and re-enter data.
- **Data quality:** Users who can't fix mistakes stop categorizing carefully.
- **Support cost:** "How do I fix a transaction?" is a predictable support ticket.

### Suggested Improvement
1. Uncomment and complete the edit flow in `TransactionColumns`.
2. Reuse the existing `TransactionExpenseForm`/`TransactionIncomeForm` pre-filled with the selected transaction's data.
3. The backend endpoint for PATCH already exists (referenced in the route action).

### Priority
**High**

---

## Missing: Export Functionality

### Problem
There is no way to export transactions, goals, or budget data. Users cannot extract their financial data.

### Business Impact
- **Trust:** Users expect data portability from financial apps — it's a trust signal.
- **Retention:** Perceived data lock-in increases churn risk.
- **Compliance:** Data export may be a regulatory requirement in some markets.

### Suggested Improvement
Add a CSV export button on the Transactions page. Minimum viable: use the existing TanStack Table's `getCoreRowModel()` to serialize visible rows to CSV client-side without a backend endpoint.

### Priority
**Medium**

---

## Reports — Only One Tab (ML Insights)

### Problem
`ReportsTab` has a single "ML Insights" tab. There are no reports for category breakdown, monthly comparison, net worth over time, or savings rate — despite components for these (`MonthComparison`, `NetWorth`, `SavingRate`, `TrendsMetric`) existing in the component directory.

### Business Impact
- **Feature discoverability:** These components exist but are unreachable in the UI.
- **Engagement:** Users have no visual summary of their financial history, reducing return visits.

### Suggested Improvement
Restore the full `ReportsTab` with tabs: Overview | Categories | Monthly Comparison | Savings Rate | Net Worth. Wire each tab to its existing component. This is likely a regression from an earlier implementation.

### Priority
**Medium**

---

# 2. UX/UI Improvements

---

## Dashboard — 5-Card Layout Breaks Grid

### Current UX Problem
`DashboardOverview` renders 5 cards in a `grid-cols-4` container. The 5th card wraps to a second row, spanning the full width and looking misaligned.

### Why It Hurts UX
Breaks visual rhythm. The dashboard is the first screen users see — broken layout signals low quality.

### Recommended UX Improvement
Change to `grid-cols-5` for 5 equal cards, or group the goal progress card into a separate row with a distinct visual treatment (e.g., full-width progress bar section below the 4 metric cards).

### Priority
**High**

---

## Transactions — Delete Button Only Visible on Hover

### Current UX Problem
The delete button in `TransactionColumns` is hidden until row hover, implemented via CSS opacity/visibility. On touch devices (mobile), there is no hover state — the delete button is inaccessible.

### Why It Hurts UX
Mobile users cannot delete transactions. Keyboard users navigating the table can't reach the delete action. This is both a usability failure and an accessibility violation.

### Recommended UX Improvement
1. Move destructive actions to a `DropdownMenu` triggered by a `...` button that is always visible and keyboard-focusable (consistent with the Goals and Budget cards).
2. Add a confirmation dialog before deletion (see Engineering section).

### Priority
**High**

---

## All Pages — Missing Loading Skeletons

### Current UX Problem
During data fetching, most pages flash null/empty content before data arrives. Only the root `Loading.tsx` component (used in the auth layout initial load) uses skeletons. Individual page sections have no loading states.

### Why It Hurts UX
Content flash (CLS - Cumulative Layout Shift) is a primary source of perceived low quality. Users on slow connections see broken layouts.

### Recommended UX Improvement
Add skeleton loaders for:
- Dashboard metric cards (`animate-pulse` rectangles matching card height)
- Budget category grid
- Goals list cards
- Reports overview cards
- Table rows (3-5 skeleton rows while fetching)

Use the existing `Skeleton` component from `app/components/ui/skeleton.tsx` — the infrastructure is already there.

### Priority
**High**

---

## Budget Detail Form — Reset Button Does Nothing

### Current UX Problem
The "Reset" button on the budget edit form (`budget-detail.tsx`) is `type="button"` with no `onClick` handler. Clicking it does nothing.

### Why It Hurts UX
Users who click Reset expecting form fields to revert to saved values find the button is broken. Broken UI elements undermine confidence in the entire product.

### Recommended UX Improvement
Either (a) wire the Reset button to refetch the budget data and re-populate form fields, or (b) remove the button entirely. A broken button is worse than no button.

### Priority
**Medium**

---

## AI Coach — No Typing Indicator / Response Loading State

### Current UX Problem
When a user sends a message to the AI Coach, there is no visual feedback that the API call is in progress. The chat is silent until the response arrives.

### Why It Hurts UX
Users don't know if their message was sent or if the system is processing. Many users will send the message again, creating duplicate requests.

### Recommended UX Improvement
1. Show a typing indicator (3 animated dots) as a temporary assistant message while the API call is pending.
2. Disable the send button and input field during the request.
3. Show a distinct visual state for failed responses (already implemented for 503, but needs consistent styling).

### Priority
**High**

---

## AI Coach — No Chat History Persistence

### Current UX Problem
Chat history in `ChatInterface` is stored in component state (`useState`). Navigating away from `/auth/ai-coach` and returning resets the conversation to the initial greeting.

### Why It Hurts UX
The core value of an AI Coach is continuity. Users building on previous questions lose all context on every navigation. This makes the feature feel unreliable and shallow.

### Recommended UX Improvement
1. Persist the `messages` array to `sessionStorage` keyed by user ID — restore on component mount.
2. Add a "Clear conversation" button so users can explicitly reset.
3. Cap stored messages at 50 to prevent storage bloat.

### Priority
**High**

---

## Navigation — No Active State Visual Clarity

### Current UX Problem
The sidebar active state uses a subtle highlight. On smaller sidebar widths, the active route is not immediately obvious.

### Why It Hurts UX
Users lose their place, especially when switching between Budgets, Goals, and Reports — which are visually similar content areas.

### Recommended UX Improvement
Add a left border accent (`border-l-4 border-primary`) on the active nav item, consistent with SaaS nav conventions. Ensure the active icon and label both use full contrast.

### Priority
**Low**

---

## Onboarding — No Help Text on Financial Input Fields

### Current UX Problem
Step 1 asks for `monthly_income`, `monthly_expenses`, `savings`, and `debt` with no explanation of what exactly to enter (gross vs net income? All debts or just monthly payments?).

### Why It Hurts UX
Ambiguous financial inputs lead to incorrect data entry, which degrades AI insights and forecast accuracy. Users with unusual financial situations (freelancers, multiple income sources) don't know how to proceed.

### Recommended UX Improvement
Add `<p>` help text or a `Tooltip` under each field explaining:
- Income: "Your average monthly take-home pay after taxes"
- Expenses: "Typical monthly spending (rent, food, bills, subscriptions)"
- Savings: "What you currently have saved (bank accounts, deposits)"
- Debt: "Total outstanding debt (loans, credit cards, not monthly payments)"

### Priority
**Medium**

---

## Forms — No Unsaved Changes Warning

### Current UX Problem
None of the edit forms (Budget Detail, Goal Update, Settings) warn users before navigating away with unsaved changes.

### Why It Hurts UX
Users who accidentally click the sidebar while editing a budget lose all their changes silently.

### Recommended UX Improvement
Use React Router v7's `useBeforeUnload` or a `blocker` to intercept navigation when a form is dirty (compare current values vs initial values). Show a confirmation dialog: "You have unsaved changes. Leave anyway?"

### Priority
**Medium**

---

## Reports — Savings Rate Card Shows Rate Without Context

### Current UX Problem
`ReportsOverview` shows Savings Rate as a percentage but no benchmark or interpretation (is 15% good? bad?).

### Why It Hurts UX
Users need context to act on a metric. A number without interpretation is noise.

### Recommended UX Improvement
Add a colored badge next to the savings rate: Green (>20% — "Excellent"), Yellow (10-20% — "Good"), Orange (5-10% — "Fair"), Red (<5% — "Needs Attention"). Include a one-line benchmark note ("Recommended: 20%+").

### Priority
**Low**

---

## Table Pagination — Resets on Tab Change

### Current UX Problem
Switching between "All" and "Monthly" tabs on the Transactions page resets the page back to 1, even if the user was on page 3.

### Why It Hurts UX
Users browsing older transactions by date filter lose their position when switching tabs.

### Recommended UX Improvement
Store the current page in URL query params per tab key (e.g., `?tab=monthly&page=3`). React Router v7 makes this clean with `useSearchParams`.

### Priority
**Low**

---

## Budget Form — No Explanation for Alert Threshold

### Current UX Problem
The budget form includes an "Alert Threshold" field with no explanation of what it does or what value is appropriate.

### Why It Hurts UX
Users won't understand whether to enter 80, 0.8, or 80% — and what triggers the alert.

### Recommended UX Improvement
Add help text: "Enter a percentage (e.g., 80) — you'll see a WARNING status when spending reaches this % of your budget limit." Validate the input range to 1-100.

### Priority
**Medium**

---

# 3. Frontend Engineering Improvements

---

## Hard-Coded Mock Data in Production Components

### Current Issue
Five components contain hard-coded arrays used as production data:
1. `DashboardGraph` — 6-month income/expense data
2. `DashboardBudgetOverview` — budget category amounts
3. `FinancialHealth` — score (78), savings rate, budget adherence, goal progress
4. `FinancialKeyInsights` — insight card content
5. `GoalsOverview` — "2 completed this year"

### Technical Risk
Any test, PR review, or screenshot looks functional but ships broken behavior to real users. There is no runtime warning that mock data is in use.

### Suggested Refactor/Improvement
1. Delete or replace mock arrays immediately. Don't leave them as "fallbacks."
2. For components pending real API wiring, render an explicit `<PlaceholderCard label="Coming soon" />` rather than fake numbers.
3. Add a `NODE_ENV === 'development'` guard with a `console.warn('MOCK DATA IN USE')` as a short-term measure during development.

### Priority
**Critical**

---

## No Error Boundaries

### Current Issue
No `ErrorBoundary` components exist at any route level. A thrown error in any component — network failure, null dereference, malformed API response — propagates to the React root and crashes the entire page.

### Technical Risk
Any API returning unexpected shape (e.g., `null` instead of `[]`) causes a white screen. This is a production reliability failure.

### Suggested Refactor/Improvement
1. Add a route-level error boundary in `app/routes/auth/layout.tsx` using React Router v7's `ErrorBoundary` export convention:
```tsx
export function ErrorBoundary() {
  const error = useRouteError();
  return <ErrorPage error={error} />;
}
```
2. Add page-level boundaries for each major section (Dashboard, Transactions, Budgets, Goals).
3. Create a shared `<ErrorCard message={} onRetry={} />` component for inline section errors.

### Priority
**Critical**

---

## Duplicate Transaction Form Components

### Current Issue
`TransactionExpenseForm` and `TransactionIncomeForm` are separate components that implement near-identical form logic — same fields, same validation schema, same API endpoint, same toast handling. The only difference is the `type` field value.

### Technical Risk
Bug fixes or validation changes must be applied in two places. This has already diverged (TanStack Form version differences are likely).

### Suggested Refactor/Improvement
Merge into a single `TransactionForm` component with a `type: "INCOME" | "EXPENSE"` prop. The `TransactionFormTab` already knows the type from which tab is active — pass it down as a prop.

### Priority
**Medium**

---

## Silent API Failures in Loader Data

### Current Issue
Multiple loaders silently handle errors by returning `null`:
```ts
// reports.tsx loader
const safeMLFetch = async (fn) => { try { return await fn(); } catch { return null; } }
```
Components then attempt to destructure `null`, causing null-reference errors or silently showing empty UI.

### Technical Risk
Users see an empty page that looks like "no data" when it's actually a backend error. No telemetry, no user feedback.

### Suggested Refactor/Improvement
1. Return structured error state: `{ data: null, error: { message, status } }`.
2. In components, check `if (error)` and render `<ErrorCard>` instead of an empty state.
3. Log errors to a telemetry service (or at minimum `console.error` in development).

### Priority
**High**

---

## No Confirmation Dialogs on Destructive Mutations

### Current Issue
Transactions can be deleted directly via a row button with no confirmation. The Goals page has a delete confirmation, but Transactions and Budget delete do not follow the same pattern.

### Technical Risk
Accidental deletion of financial data with no recovery mechanism damages user trust and creates support escalations.

### Suggested Refactor/Improvement
Create a shared `<ConfirmDeleteDialog>` component:
```tsx
<ConfirmDeleteDialog
  title="Delete Transaction?"
  description="This action cannot be undone."
  onConfirm={() => deleteTransaction(id)}
/>
```
Apply consistently to: Transaction delete, Budget delete, Goal delete, Account deletion.

### Priority
**High**

---

## DataTable Hard-Coded Page Size

### Current Issue
`DataTable.tsx` sets page size to 10 items and ignores any `pageSize` prop:
```ts
const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 10 });
```

### Technical Risk
Different tables (transactions, goals) have different appropriate page sizes. This cannot be customized without modifying the shared component directly, coupling unrelated features.

### Suggested Refactor/Improvement
Add a `defaultPageSize?: number` prop defaulting to 10. Pass it to `useState` initializer. Optionally add a page size selector dropdown for tables with many rows.

### Priority
**Low**

---

## Token Passed as Prop Through Deep Component Trees

### Current Issue
The JWT token is extracted in the route loader and passed as a prop through multiple component layers to reach hooks and fetch functions. Example: `transactions.tsx` → `TransactionTable` → `useGetTransaction` hook.

### Technical Risk
Any new component needing the token must update every intermediate component's prop interface — this is "prop drilling" and will become unmanageable as the component tree grows.

### Suggested Refactor/Improvement
Create a `TokenContext` that wraps the auth layout:
```tsx
// In auth/layout.tsx
<TokenContext.Provider value={token}>
  <Outlet />
</TokenContext.Provider>
```
Use `useToken()` in any component or hook that needs it. Remove token props from all component interfaces.

### Priority
**Medium**

---

## Missing React.Suspense / Code Splitting

### Current Issue
All route components are eagerly bundled. Recharts (a large dependency) is loaded even on pages that don't use charts. No `React.lazy()` usage observed.

### Technical Risk
Initial bundle size is larger than necessary. Users on slow connections experience longer Time-to-Interactive.

### Suggested Refactor/Improvement
1. Wrap chart-heavy components (`PieChart`, `DashboardGraph`, `ForecastChart`) with `React.lazy()` and `<Suspense fallback={<ChartSkeleton />}>`.
2. React Router v7 supports route-level code splitting natively — enable it for heavy routes like Forecast and Reports.

### Priority
**Low**

---

## Redux Toolkit Installed But Unclear Usage

### Current Issue
`@reduxjs/toolkit` is listed in `package.json` but no Redux store, slices, or `Provider` wrappers are visible in the codebase. TanStack Query handles all server state.

### Technical Risk
Unused dependencies increase bundle size and create confusion for new contributors.

### Suggested Refactor/Improvement
Audit whether RTK is actually used anywhere (`grep -r "useSelector\|useDispatch\|createSlice" app/`). If unused, remove it from `package.json` to reduce bundle weight.

### Priority
**Low**

---

## Overview Component Color Mapping Is Fragile

### Current Issue
`Overview.tsx` maps string color names to Tailwind classes via a switch statement (5 colors). Any caller passing an unlisted color silently falls through to a default.

### Technical Risk
New feature areas that need different colors (e.g., purple for AI metrics) require modifying the shared component, coupling unrelated concerns.

### Suggested Refactor/Improvement
Accept Tailwind class strings directly:
```tsx
// Instead of: color="green"
// Use: colorClass="bg-green-100 text-green-700"
```
Or extend to a typed `ColorVariant` union. This removes the mapping layer entirely.

### Priority
**Low**

---

# 4. Accessibility Review

---

## Delete Buttons Not Keyboard-Accessible

**Issue:** Transaction row delete buttons use hover-only visibility (`opacity-0 group-hover:opacity-100`). Keyboard users navigating with Tab cannot reach these buttons.

**Fix:** Move to always-visible `DropdownMenu` (see UX section). At minimum, ensure delete buttons receive focus on Tab navigation even when visually hidden.

**WCAG Level:** A (Violation)

---

## Icon-Only Buttons Missing ARIA Labels

**Issue:** Several icon-only buttons throughout the app (sidebar icons in collapsed state, action icons in table rows) lack `aria-label` attributes. Screen readers announce them as generic "button."

**Fix:** Add `aria-label="Delete transaction"`, `aria-label="Edit budget"` etc. to every icon-only interactive element.

**WCAG Level:** A (Violation)

---

## Form Field Errors Inconsistent Across Pages

**Issue:** Registration and Goal Update forms correctly use `aria-invalid="true"` and `aria-describedby` linking to error messages. Budget forms, transaction forms, and onboarding Step 2 do not follow this pattern consistently.

**Fix:** Standardize all form error patterns. Create a `<FormField>` wrapper that automatically applies aria attributes when an `error` prop is present. The existing `field.tsx` primitive in `app/components/ui/` can be extended for this.

**WCAG Level:** AA (Violation)

---

## Modal Focus Management

**Issue:** When dialogs open (Add Transaction, Add Budget, etc.), focus is not explicitly moved to the first interactive element inside the dialog. Users pressing Tab from the trigger button may focus on elements behind the modal overlay.

**Fix:** Radix UI `Dialog` component handles this automatically when used correctly. Verify that the `Modal.tsx` wrapper passes `autoFocus` to the first input, and that `DialogContent` is not wrapped in a way that blocks Radix's built-in focus trap.

**WCAG Level:** AA (Violation)

---

## Color Contrast — Status Badges

**Issue:** `SAFE` status badge (green on green-tinted background) and `WARNING` badge (orange on light orange) may not meet 4.5:1 contrast ratio for normal text, especially with Tailwind's default color scale.

**Fix:** Test badge combinations with a contrast checker. Consider using outlined badges (colored border + dark text) instead of filled variants for status indicators. Add a non-color indicator (icon or text prefix) so color-blind users can distinguish statuses.

**WCAG Level:** AA (Violation)

---

## Sidebar — Logo and Brand Name Not Focusable

**Issue:** The app logo/brand link in the sidebar is not a focusable element and cannot be reached via keyboard. There is no "Skip to main content" link.

**Fix:**
1. Wrap the brand mark in an `<a href="/auth">` with `aria-label="Go to Dashboard"`.
2. Add a visually-hidden skip link at the top of every page: `<a href="#main-content" className="sr-only focus:not-sr-only">Skip to main content</a>`.

**WCAG Level:** A (Violation)

---

## Calendar / DatePicker — Keyboard Navigation

**Issue:** The `DatePicker` component opens via `ArrowDown` key (good), but the calendar inside may not follow the ARIA Date Picker design pattern for full keyboard navigation (arrow keys to navigate dates, Enter to select, Escape to close).

**Fix:** Test with keyboard-only navigation. `react-day-picker` (used under the hood) supports this pattern natively — verify the wrapper doesn't intercept keyboard events before they reach the calendar.

**WCAG Level:** AA

---

## Checkbox Group — Onboarding Goals Selection

**Issue:** In Onboarding Step 2, multiple checkboxes for financial goals lack a wrapping `<fieldset>` and `<legend>` to group them semantically.

**Fix:**
```html
<fieldset>
  <legend>What are your financial goals?</legend>
  <!-- checkboxes -->
</fieldset>
```

**WCAG Level:** A

---

# 5. Mobile Experience Review

---

## Dashboard Cards — No Mobile Breakpoints

**Issue:** Dashboard uses `grid-cols-4` with no responsive variant. On screens under 768px, cards are either too small to read or overflow horizontally.

**Fix:** Apply responsive grid classes: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-4`. Cards should stack on mobile and show 2 columns on tablet.

**Pages Affected:** Dashboard, Budget Overview, Reports Overview

---

## Transaction Table — Overflows on Mobile

**Issue:** The transactions table has multiple columns (Description, Category, Date, Type, Amount) that don't fit on a 375px screen. There is no horizontal scroll hint or column hiding on small screens.

**Fix:**
1. On mobile, hide secondary columns (Category, Date) and show them in an expanded row or detail sheet.
2. Alternatively, switch to a card list view below `md` breakpoint using a responsive component swap.
3. Minimum: add `overflow-x-auto` to the table container so users can scroll horizontally.

---

## Forms — Input Fields Too Narrow on Mobile

**Issue:** Several form layouts use side-by-side input pairs (`grid-cols-2`) without responsive fallback to single-column on small screens.

**Fix:** Add `grid-cols-1 md:grid-cols-2` to all multi-column form grids. Verify touch targets meet 44×44px minimum.

---

## Sidebar — No Mobile Bottom Navigation

**Issue:** The app uses a sidebar that collapses to a sheet on mobile. There is no bottom tab bar for mobile, which is the standard navigation pattern on touch devices. The hamburger menu approach adds friction for common actions.

**Fix:** Consider adding a bottom navigation bar for the 4-5 most used sections (Dashboard, Transactions, Budgets, Goals, AI Coach) visible only on mobile (`md:hidden`). Keep the sidebar for desktop.

---

## Modal Dialogs — No Full-Screen Mobile Variant

**Issue:** Modals (Add Transaction, Add Budget, Add Goal) open as centered dialogs that may be too small or partially off-screen on mobile. Long forms inside modals require scrolling within the modal.

**Fix:** Use Radix UI's `Sheet` component on mobile to render forms as a bottom drawer rather than a centered dialog. Apply conditionally based on screen width: `useMediaQuery('(max-width: 768px)')`.

---

## Sticky Header Missing on Long Lists

**Issue:** On mobile, the page header with action buttons (e.g., "Add Transaction") scrolls out of view when browsing a long list. Users must scroll back to the top to add new data.

**Fix:** Make the page header sticky (`sticky top-0 z-10 bg-background`) on mobile. The `AppSidebar` header is already sticky — extend this pattern to page-level headers.

---

# 6. Quick Wins

| Improvement | Estimated Effort | Expected Impact |
|-------------|-----------------|-----------------|
| Fix sidebar hardcoded user name — pull from JWT payload | 1 hour | High — trust |
| Remove/replace static mock data in 5 components | 4-8 hours | Critical — trust |
| Add `aria-label` to all icon-only buttons | 2 hours | Medium — accessibility |
| Fix broken Reset button in budget-detail form | 30 min | Low — UX polish |
| Add `overflow-x-auto` to transaction table | 15 min | Medium — mobile |
| Add `ConfirmDeleteDialog` to transaction delete | 2 hours | High — data safety |
| Persist AI chat history in sessionStorage | 2 hours | High — engagement |
| Add typing indicator to AI Coach chat | 1 hour | Medium — UX polish |
| Fix 5-card grid layout in DashboardOverview | 30 min | Medium — visual quality |
| Add responsive grid breakpoints to Dashboard/Budget cards | 2 hours | High — mobile |
| Wire `GoalsOverview` "completed" count to real data | 1 hour | Medium — trust |
| Add `ErrorBoundary` export to auth layout | 2 hours | High — reliability |
| Add loading skeletons to Budget and Goals pages | 3 hours | Medium — perceived performance |
| Add "Skip to main content" link | 30 min | Low — accessibility |
| Remove or audit unused Redux Toolkit dependency | 30 min | Low — bundle size |

---

# 7. High Impact Features Missing

---

## Notifications System

**What:** In-app and push notifications for budget threshold alerts (WARNING/EXCEEDED), goal deadline reminders, and unusual spending patterns.

**Why it matters:** Budget alerts are the #1 feature users expect from a budgeting app. The `alert_threshold` field exists in the data model — there is no notification delivery. The Settings page has Notifications UI but it appears non-functional.

**Implementation path:** Notification preferences in Settings → backend push via Web Push API or email → in-app notification bell with unread count badge.

---

## Transaction Import (CSV/Bank Statement)

**What:** Allow users to bulk-import transactions from a CSV file (common bank export format).

**Why it matters:** Manual transaction entry is the #1 reason users abandon budgeting apps. Lowering the entry friction directly improves activation and retention.

**Implementation path:** File upload on Transactions page → client-side CSV parsing → review/map columns → bulk POST to `/auth/v1/transactions`.

---

## Monthly Budget Reset & Historical Tracking

**What:** Automatic monthly budget usage reset with history preserved for comparison.

**Why it matters:** The current budget system appears to track cumulative spending but users expect monthly resets aligned with their pay cycle.

**Implementation path:** Add `period_start`/`period_end` to budget usage. Show last month's performance vs. current month in BudgetBreakdown.

---

## Keyboard Shortcuts

**What:** Common keyboard shortcuts: `T` → New Transaction, `B` → New Budget, `G` → New Goal, `/` → Focus search/filter.

**Why it matters:** Power users who do daily expense tracking benefit enormously. It signals product maturity.

**Implementation path:** Global `keydown` listener in auth layout. Show shortcut hints in the UI on hover (Tooltip with shortcut key).

---

## Activity / Audit Log

**What:** A read-only log of all mutations (transaction created/deleted, budget updated, goal contribution added).

**Why it matters:** Users need audit trails for financial data. Required for financial compliance in many markets. Also catches "I didn't delete that" support issues.

---

## Dark Mode

**What:** Theme toggle with `dark:` Tailwind classes applied app-wide.

**Why it matters:** `next-themes` is already installed. Dark mode is a hygiene feature users expect in 2026.

**Implementation path:** Add a theme toggle button in Settings → Account tab. Apply `dark:` variants to existing Tailwind classes. Most Radix UI components handle this automatically.

---

## Recurring Transactions

**What:** Mark a transaction as recurring (weekly/monthly/yearly) and auto-generate future instances.

**Why it matters:** Salaries, rent, subscriptions — most financial activity is recurring. Manual re-entry every month is the second biggest friction point after initial import.

---

## Spending Alerts / Anomaly Notifications

**What:** When the ML anomaly endpoint detects unusual spending (already exists at `/auth/v1/ml/anomaly`), surface this as an in-app alert banner.

**Why it matters:** The ML infrastructure is already built. Surfacing its output to users creates immediate perceived value from the AI features.

**Implementation path:** Check anomaly endpoint result in the Dashboard loader. If anomalies exist, render an `AlertBanner` component above the dashboard cards.

---

# 8. Final Recommendations

---

## Top 5 Highest Priority Improvements

1. **Remove all hardcoded mock data** — Wire `DashboardGraph`, `DashboardBudgetOverview`, `FinancialHealth`, `FinancialKeyInsights`, and `GoalsOverview` to real API data. This is a credibility failure that blocks production launch.

2. **Fix sidebar hardcoded user identity** — "John Doe / johndoe@email.com" in a financial app is a P0 trust issue. Pull name and email from the decoded JWT payload or a `/auth/v1/user` endpoint.

3. **Add error boundaries** — Deploy route-level `ErrorBoundary` components to prevent full-page crashes from API failures. This is the most impactful reliability improvement per line of code.

4. **Add transaction delete confirmation** — A destructive, irreversible action on financial data with no confirmation is a user safety issue. Implement `ConfirmDeleteDialog` consistently across all delete actions.

5. **Fix 5-card grid layout and add responsive breakpoints** — The broken dashboard layout is the first thing every user sees. It signals poor quality before users even interact with a feature.

---

## Top UX Fixes

- Persist AI Coach chat in `sessionStorage` (feature becomes useless without continuity)
- Add loading skeletons to Budget, Goals, and Reports pages
- Make transaction delete/edit accessible on mobile (remove hover-only pattern)
- Add empty states with CTAs on Dashboard and Transactions for new users
- Fix broken Reset button on budget-detail form
- Add help text to onboarding financial input fields

---

## Top Business Opportunities

- **Notifications** — deliver budget alerts via the existing `alert_threshold` model field; highest user-expected feature
- **CSV import** — removes the #1 friction point for new user activation
- **Dark mode** — `next-themes` is installed; estimated 2-day implementation; high user demand
- **Anomaly alert banner** — ML anomaly detection is already built; surface it on Dashboard for immediate AI value demonstration
- **Restore full Reports tabs** — `MonthComparison`, `NetWorth`, `SavingRate` components exist but are unreachable; restoring them adds significant perceived value with minimal engineering work

---

## Technical Debt to Fix First

1. **Mock data removal** — Blocks any meaningful QA or user testing
2. **Error boundaries** — Blocks production reliability
3. **Token prop drilling** — Replace with `TokenContext` before adding more features (complexity grows exponentially)
4. **Duplicate transaction forms** — Merge before adding edit functionality (bug surface doubles otherwise)
5. **DataTable hard-coded page size** — Fix before adding new tables (growing technical constraint)
6. **Audit and remove Redux Toolkit** — Dead dependency that bloats the bundle and confuses contributors

---

*This document was generated from a full codebase audit of the `feat/budgets-integration` branch as of 2026-05-20. It should be reviewed and updated after each major feature release.*
