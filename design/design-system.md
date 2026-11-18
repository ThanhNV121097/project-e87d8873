# Design System — Đèo Bikes

> Source of truth: the approved design in `code/frontend/src` — `theme.css` (every colour, face, radius,
> space and motion value), `index.css` (how Tailwind names them), `content.json` (every word),
> `App.tsx` and `components/` (the screens). Every value below is extracted from it. Changing a value
> here without changing the approved design is a defect.
>
> The design is the template **Pastel** as the owner chose it (`design/template.jpg`), filled for a
> mountain-bike shop (`design/moodboard.md`): a white page, one near-black navy for the type and for
> everything pressed, and the only colour on the page in the **cards** — four pale tints a thing stands on.
> One face (Inter); a heading's emphasis is a change of **weight inside the line**, never a second face
> or a colour.

Last updated: 2026-11-18

## 0. How tokens are named and used

- Tokens are CSS custom properties in `src/theme.css`. Their names are already semantic (`--ink`,
  `--accent`, `--tint-1`), so this document uses **those names**, not renamed copies.
- `src/index.css` (`@theme inline`) gives each a Tailwind utility: `bg-ground`, `bg-surface`, `text-ink`,
  `text-ink-soft`, `bg-accent`, `text-accent-ink`, `border-line`, `border-field`, `text-danger`,
  `bg-wash`, `font-display`, `font-body`, `rounded`, `rounded-sm`, `max-w-page`. A story writes these
  names, never a raw `#hex`, `rgb()` or a palette colour from outside the theme.
- The four tints, `--accent-soft`, `--glass`, `--space-*` and `--shadow-*` have **no utility name**; the
  design reaches them as `bg-[var(--tint-1)]`, `bg-[var(--glass)]`, `shadow-[var(--shadow-sm)]`,
  `gap-[var(--space-6)]`. That form names a token and is the only arbitrary-value form that is allowed.
- `src/theme.css` and `src/index.css` belong to the scaffold. A story never edits either.
- One mode: `color-scheme: light`. There is no dark palette and no theme switch.

## 1. Foundations

### 1.1 Color

#### Storefront (every page except `/admin`)

| Token | Value | Used for |
|---|---|---|
| `--ground` | `#FFFFFF` | Page background; field fill; the white pill on a tint; the round save button |
| `--surface` | `#F6F6F6` | The grey band (catalogue head, "where to start", bag head, "ask the shop" band); Select trigger and list; the Zalo dock button |
| `--ink` | `#191A2E` | All type; outlined button edge and label; selected-pill fill |
| `--ink-soft` | `#5F6373` | Secondary text, captions, eyebrows, struck-through old price, disabled labels |
| `--accent` | `#191A2E` | Primary fill — every solid button, the chosen chip, the footer card, the navy bands, the bag count dot. Same value as `--ink` on purpose: "the ink that presses every button also closes every page" |
| `--accent-ink` | `#FFFFFF` | Text and icons on `--accent` (and, at 75–80 % alpha, secondary text on the footer and navy bands) |
| `--accent-soft` | `#FFEBED` | The "saved" amount pill on a bike's price |
| `--danger` | `#B42318` | What is wrong or lost only: the out-of-stock mark, an order error line. Never brand colour |
| `--danger-ink` | `#FFFFFF` | Text on `--danger` |
| `--wash` | `color-mix(in srgb, var(--ink) 6%, var(--surface))` ≈ `#E9E9EA` | Quiet hover fill: icon buttons, chips, the remove button |
| `--line` | `rgba(25,26,46,0.12)` ≈ `#E3E4E6` on white | Hairlines: header rules, dividers, card edges, thumbnails. Also Tailwind's default `border-color` |
| `--field-line` | `rgba(25,26,46,0.45)` ≈ `#9898A1` on white | The edge of anything you operate: inputs, textarea, Select, size pills, option cards, checkboxes |
| `--tint-1` | `#FFEBED` | Pink card; bag summary; empty states; chosen option card |
| `--tint-2` | `#F6FFB0` | Butter card; reply-time pill; order-button panel; "now" step on the order receipt |
| `--tint-3` | `#EFE2FE` | Lilac card; the help band under the catalogue |
| `--tint-4` | `#FFEFE2` | Peach card; the three-promises band |
| `--glass` | `rgba(255,255,255,0.78)` | The bar across a tinted card's foot and the bag's total row (with `backdrop-blur-md` on the card) |
| `--product-ground` | `#FFFFFF` | The tile behind every product photograph until it loads |
| `--product-ground-ink` | `#5F6373` | Caption words on that tile when a bike has no photograph |
| `--picture-ground` / `--picture-ink` | `var(--surface)` / `var(--ink-soft)` | The flat block standing where a photograph is not there yet (`ProductImage` re-points both at the white tile) |

Tint assignment: a tinted card takes `TINTS[i % 4]` (pink, butter, lilac, peach) by its place in its
row; a bike's own page uses the tint of its riding type's position in the list. Tints are pale on
purpose; text on them is `--ink` or `--ink-soft` only.

#### Management area (`/admin`, scoped by `[data-dashboard]`)

The public site is the owner's brand; the management area is a work tool with ONE fixed look that a
design does not restyle. It redefines the same token names inside `[data-dashboard]`:

| Token | Value | Token | Value |
|---|---|---|---|
| `--ground` | `#F8FAFC` | `--accent` | `#2563EB` |
| `--surface` | `#FFFFFF` | `--accent-ink` | `#FFFFFF` |
| `--ink` | `#0F172A` | `--line` | `#E2E8F0` |
| `--ink-soft` | `#64748B` | `--field-line` | `#CBD5E1` |
| `--danger` | `#DC2626` | `--wash` | `#F1F5F9` |
| `--radius` | `6px` | `--radius-sm` | `4px` |

It keeps the owner's typeface, uses a denser spacing, and its controls are the shadcn/ui kit in
`src/components/ui`. It is not a second mode: nothing a visitor can switch reaches it.

#### Contrast audit

Ratios are WCAG relative-luminance contrast, computed from the values above (alpha colours flattened
on the ground they are drawn on). Body text needs 4.5:1, large text (≥ 24 px, or ≥ 18.66 px bold) 3:1,
a control's edge 3:1.

| Foreground | Background | Ratio | Passes |
|---|---|---|---|
| `--ink` | `--ground` | 17.1:1 | AA, AAA |
| `--ink` | `--surface` | 15.8:1 | AA, AAA |
| `--ink` | `--wash` | 14.1:1 | AA, AAA |
| `--ink` | `--tint-1` | 14.9:1 | AA, AAA |
| `--ink` | `--tint-2` | 16.2:1 | AA, AAA |
| `--ink` | `--tint-3` | 13.8:1 | AA, AAA |
| `--ink` | `--tint-4` | 15.2:1 | AA, AAA |
| `--ink-soft` | `--ground` | 6.0:1 | AA |
| `--ink-soft` | `--surface` | 5.5:1 | AA |
| `--ink-soft` | `--tint-1` | 5.2:1 | AA |
| `--ink-soft` | `--tint-2` | 5.6:1 | AA |
| `--ink-soft` | `--tint-3` | 4.8:1 | AA (the tightest text pair on the site) |
| `--ink-soft` | `--tint-4` | 5.3:1 | AA |
| `--accent-ink` | `--accent` | 17.1:1 | AA, AAA |
| `--accent-ink` at 75 % | `--accent` | 10.0:1 | AA, AAA (lowest alpha used: footer and navy bands) |
| `--danger-ink` | `--danger` | 6.6:1 | AA |
| `--danger` | `--ground` | 6.6:1 | AA (out-of-stock mark, order error line) |
| `--accent` | `--ground` | 17.1:1 | AA, AAA ("View all" link text) |
| `--field-line` | `--ground` | 2.9:1 | **FAIL** against the 3:1 edge rule — see Known deviations |
| `--line` | `--ground` | 1.3:1 | Not a control edge — decorative hairline only |
| Management: `--ink-soft` `#64748B` | `--ground` `#F8FAFC` | 4.5:1 | AA |
| Management: `--accent-ink` | `--accent` `#2563EB` | 5.2:1 | AA |

### 1.2 Spacing

Base unit: `4px`. The design's own scale, in `theme.css`:

| Token | Value |
|---|---|
| `--space-1` | `4px` |
| `--space-2` | `8px` |
| `--space-3` | `12px` |
| `--space-4` | `16px` |
| `--space-6` | `24px` |
| `--space-8` | `32px` |
| `--space-12` | `48px` |
| `--space-16` | `64px` |

Tailwind's spacing utilities (`p-4`, `gap-6`, `mt-12`) are the same 4 px grid, so `p-4` is `--space-4`.
The design also uses steps between and above the named tokens — `1.5` (6 px), `2.5` (10 px), `3.5`
(14 px), `5` (20 px), `7` (28 px), `9` (36 px), `10` (40 px), `14` (56 px), `20` (80 px) and `24`
(96 px) — all on the same 4 px grid.

Rhythm by role (what the design actually does):

| Where | Value |
|---|---|
| Between home-page sections, and before the product page's specification, ask band and related row | `pt-24` (96 px) |
| After the hero, before the three tinted cards | `pt-16` (64 px) |
| Inside a card (tint card, review card, summary, promise box) | `p-5` to `p-6` (20–24 px); the large tinted bands `px-6 py-10` to `lg:p-12` |
| Grid gap between cards | `var(--space-6)` (24 px) |
| A form's stacked fields | `gap-4` (16 px); groups of a form `gap-10` (40 px) |
| Between a card's lines of text | `mt-1` to `mt-4` (4–16 px) |
| The foot of every page above the footer | `pb-8` (32 px); the footer is `mt-24` |

### 1.3 Typography

One face, no second. Inter, loaded as a variable font from Google Fonts in `index.html`
(`family=Inter:ital,opsz,wght@0,14..32,100..900`, `display=swap`, with `preconnect`).

- Body: `var(--font-body)` = `"Inter"`, then `system-ui, sans-serif` — `font-body`
- Headings: `var(--font-display)` = `"Inter"`, then `system-ui, sans-serif` — `font-display`
- Mono: none; the design draws no monospaced text. Figures use `tabular-nums` so prices and counts align.

The page body is `17px / 1.6`, weight 400, `-webkit-font-smoothing: antialiased`.

The size ramp the design actually uses (Tailwind utilities, except the two tokens):

| Token / utility | Size | Weight | Line height | Used for |
|---|---|---|---|---|
| `--size-h1` | `clamp(2.25rem, 3.9vw, 3.5rem)` = 36 px → 56 px (56 px at 1440) | 300, with a 700 run | `1.1`–`1.12` | The one `h1` of a page; the Home "Bike of the month" heading |
| `--size-h2` | `clamp(1.625rem, 2.4vw, 2.125rem)` = 26 px → 34 px | 300, with a 700 run | `leading-tight` (1.25), `1.08` on a bike's own name | Section headings; the bike's name in its tinted head |
| `text-4xl` | 36 px | 700, `tracking-tight` | 1.1 | A bike's price on its page |
| `text-[1.75rem]` / `lg:text-[2rem]` | 28 px / 32 px | 300 brand + 700 model | `1.08` | Model name on a full tinted card |
| `text-2xl` | 24 px | 700 (stat) or 400 (empty-state line) | 1.33 | Counts, order number, empty-state headline, mobile menu links at `text-3xl` (30 px, weight 300) |
| `text-xl` | 20 px | 400–700 | 1.4 | Compact tinted card name, order total, help-band line |
| `text-lg` | 18 px | 400–600 | 1.55 | Cart title, form group legend, search field, desktop departments wordmarks |
| `text-base` | 16 px | 600 name / 700 price | 1.5 | Card name and price |
| `text-sm` | 14 px | 400 / 500 / 600 | 1.43 | Secondary body, nav links, every button label, form labels |
| `text-xs` | 12 px | 400 / 500 | 1.33 | Eyebrows (uppercase), captions, stock line, review footer, filter group titles |
| `text-[0.9375rem]` | 15 px | 400 | `leading-relaxed` | The bag's lead paragraph |
| `text-[0.6875rem]` | 11 px | 400 | — | Rider height under a frame-size pill |
| `text-[0.625rem]` | 10 px | 400 | — | The bag count in the header dot |

Heading levels are used in order: one `h1` per page, `h2` for sections, no skipping. Footer column
titles are paragraphs, not headings.

| Token | Value | Used for |
|---|---|---|
| `--weight-display` | `300` | h1–h3 by default (set in `index.css`, `@layer base`) |
| Body weight | `400` | Running text |
| `font-medium` | `500` | Buttons, chosen chip, small links |
| `font-semibold` | `600` | Card names, active nav link, labels, stock "In stock" word |
| `font-bold` | `700` | Prices, counts, the **bold run** of a heading |
| `--tracking-display` | `-0.03em` | h1–h3 by default; the Home h1 sets `-0.01em`; tinted-card names and prices set `tracking-tight` (-0.025em) |
| Eyebrow tracking | `0.14em` uppercase | Brand · riding type above a card name (`text-xs uppercase`) |
| Wordmark tracking | `0.2em` uppercase | The departments row on Home (`text-sm`, `lg:text-lg`, weight 600) |
| `tracking-wider` | `0.05em` | Filter group titles, size-guide table head |

**The one typographic device — a heading that changes weight inside the line.** A light run (300), a bold
run (700), a light tail, in the same size and colour: "Built for / **Passes. Trails** / & rough ground".
It is the `Weights` component (`k.light`, `k.bold`, `k.tail`) or two spans (`font-light` + `font-bold`)
where the words are the record's own. Preflight makes headings inherit weight, so a heading's weight is
always stated by this token or by the span, never assumed.

### 1.4 Radius, border, shadow, motion

| Token | Value | Used for |
|---|---|---|
| `--radius` (`rounded`) | `24px` | Cards, product tiles, bands, review cards, summary panels, the Select trigger and list |
| `--radius-sm` (`rounded-sm`, `rounded-[var(--radius-sm)]`) | `14px` | Form fields, frame-size pills, the photograph window inside a tinted card, gallery thumbnails |
| `rounded-full` | `9999px` | Every solid and outlined button, chips, white pills on a tint, round icon buttons, the heart, the dock, the bag count dot, avatars of the proof row |
| `--radius-pill` | `999px` | The mobile menu toggle (`rounded-[var(--radius-pill)]`); same job as `rounded-full` |
| Border width | `1px` (`border`) | Hairlines (`border-line`), field and option edges (`border-field`), outlined buttons (`border-ink`); a tab's active underline is `2px` |
| `--shadow-sm` | `0 1px 2px rgba(27,25,23,0.06)` | The heart on a card; the glass price bar on a tinted card |
| `--duration-fast` | `120ms` | Defined for hover and focus changes |
| `--duration-base` | `220ms` | Defined for panel changes |
| `--ease-out` | `cubic-bezier(0.2, 0.7, 0.2, 1)` | The one easing; the product picture's zoom uses this exact curve |

Resting cards have **no shadow** — a card is a tint or a hairline edge, not a floating object. Floating
layers (Select list, ContactDock, CartDrawer) use Tailwind's `shadow-lg` / `shadow-xl`, which are not
tokens (see Known deviations). `--shadow-md` is defined in `theme.css` and no storefront element uses it.

**Motion — level "lively", the owner's choice, in `src/site.json`.** It comes from `src/motion` and is
not written by hand in a story: the hero heading's words rise (`SplitHeading`), the hero pill follows
the pointer (`data-magnetic`), the in-stock count counts up (`data-count`), the three tinted cards tilt
(`data-tilt`), their photographs drift against the scroll (`motion="parallax"` on `Picture`), a grid
card's picture grows 5 % and swaps to its second photograph under the pointer (`0.6s` zoom, `0.35s`
swap), and a bike's main picture is looked into at twice its size under the pointer (`data-lens`).

Motion respects `prefers-reduced-motion: reduce`: `movement()` returns `still`, nothing moves, and
nothing is hidden waiting to move — every device starts from the finished page.

### 1.5 Layout and breakpoints

Tailwind's default breakpoints; the design uses three of them and no `xl`/`2xl`.

| Name | Min width | Container | Columns | Gutter |
|---|---|---|---|---|
| base (phone) | 0 | `max-w-page` (1284 px), padded by `--gutter` | 1 for bikes, reviews, promises; 2 for small things (`Grid phone`) | `--gutter` = `clamp(20px, 5.5vw, 72px)` page edge; 24 px between cards |
| `sm` | 640px | same | up to 3 across (`Grid` tablet) | same |
| `md` | 768px | same | the three-up bands (promises, contact reasons) | same |
| `lg` | 1024px | same | up to `Grid max` across (3 on Home, Shop, related row); 12-column rows for the hero, footer and contact; the filter rail sits beside the grid | same |

`lg` is the switch between the phone design and the desktop design: the nav, search field and sign-in
icon appear, the menu toggle and `ContactDock` go, the product page becomes picture-left / buy-right,
and filters stop folding.

Fixed dimensions: header row `--header-h` = `68px` between two hairlines; the mobile menu panel starts
at that line; `--dock-space` = `88px + env(safe-area-inset-bottom)` kept clear at the foot of the body
below 1024 px while the dock shows.

Pictures: bikes are landscape — every product tile is `4/3` (the gallery, cards, bag lines); `1/1` for
the 20 px-square cart thumbnails and the proof rounds; `4/5` exists in `ProductImage` and is not used.
**Bleed**: a hero photograph, a band or a form panel may run off the screen's edge while its content
stays on the page's own edge (`Bleed`, and the `-mx-[var(--gutter)]` / `lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))]`
pattern).

Z-index scale (only these values are used):

| Layer | Value |
|---|---|
| Base | `0` |
| Sticky header (management area only; the storefront header is not sticky) | `10` |
| Dropdown (Select list) | `20` |
| Overlay, cart drawer, ContactDock, motion progress line and cursor | `50` |
| Mobile menu and mobile search panel (above the dock) | `60` |

## 2. Components

The storefront's parts live in `src/components/`; the design draws each state listed and no other. Where
a state is not listed, the design does not draw it — a later story must not invent it. Where Focus is not
listed, the design draws **no focus style of its own** and the browser's default focus ring applies (see
Known deviations); nothing in the storefront may remove that ring without a replacement.

### 2.1 Button (a class recipe, not a component)

**Purpose** — the action of a screen. One solid button per region; the outlined one is the quieter
second choice. Never a rectangle: every button is a pill.

**Anatomy** — `[icon?] [label] [arrow icon?]`; the arrow (`Icon arrow`, 16 px) follows a label that moves
the visitor on.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Primary | `rounded-full bg-accent text-accent-ink font-medium` | The region's main action: "See all bikes", "Add to bag", "Check out", "Place order", "Send message", "Ask the shop to call" |
| Outline | `rounded-full border border-ink text-ink` | The second choice: "Choose a size" on a card, "Ask about this bike" beside "Add to bag", "Show more" |
| Inverted | `rounded-full bg-ground text-ink` | On a navy band only ("Browse" on the empty receipt) |
| Round icon | `h-10 w-10 rounded-full` + `hover:bg-wash` (header, remove-from-bag) or `bg-accent text-accent-ink` (the arrow on a tinted card) | An action with no word: search, bag, sign-in, remove, "choose a size" on a tinted card |

**Sizes** (by padding; the label is `text-sm` unless the button is the product page's, which is 16 px)

| Size | Height | Padding | Used by |
|---|---|---|---|
| Hero | ~48 px | `px-9 py-3.5` | The pill on the hero headline's last line |
| Large | ~52 px | `px-6`–`px-8 py-3.5` | Add to bag, Ask about this bike, Check out, Place order |
| Standard | ~44 px | `px-7 py-3` | "See the bike", "Send message", "Ask the shop to call", empty-bag browse |
| Small | ~40 px | `px-5`–`px-6 py-2`–`py-2.5` | Header "Search", help-band link, "Clear all filters" |
| Round | 40 px | `h-10 w-10` | Icon buttons |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | As the variant above | — |
| Hover | Primary: `hover:opacity-90` (hero pill, month CTA, Add to bag, Check out, the tinted card's arrow). Outline: fills navy (`hover:bg-accent hover:text-accent-ink`) on cards, chips on Home and "Show more"; fills `--wash` (`hover:bg-wash`) on the product page's "Ask" and the Shop chips. Round icon: `--wash`. Some primary buttons draw no hover (header "Search", Send message, Ask the shop to call, Place order) | `--accent`, `--wash` |
| Disabled | `disabled:opacity-60` on a submitting form's button, with the label changed to "Sending…"; a sold-out size pill is its own component (2.8) | — |

**Accessibility** — a `<button>` for an action, a `<Link>` for a navigation; an icon-only control always
carries an `aria-label` (the words come from `content.json`). The header's icon buttons are 40×40 and
the heart is 32×32, under the 44×44 minimum (see Known deviations).

### 2.2 Chip (riding-type filter)

**Purpose** — choose a riding type (Hardtail, Cross-country, Trail, E-MTB…) on Shop (a button) or jump to
it from Home and an empty bag (a link).

**Anatomy** — `[label]`; `rounded-full px-5 py-2 text-sm` on Shop, `px-4 py-2 text-xs` on Home.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | `border border-ink text-ink` | `--ink` |
| Hover | Shop: `hover:bg-wash`. Home and bag: `hover:bg-accent hover:text-accent-ink` | `--wash`, `--accent` |
| Selected | Solid navy, `font-medium`; `aria-pressed="true"` on Shop; the "All" chip on Home is drawn selected | `--accent`, `--accent-ink` |

**Accessibility** — the group is a `<nav aria-label="Riding types">`; the selected one is conveyed by
`aria-pressed`, not colour alone.

### 2.3 Pill label

**Purpose** — a small fact on a surface: the riding type on a tinted card, "You save …", "A person
answers within 2 hours". Never an action.

**Anatomy** — `[icon?] [text]` in `rounded-full px-3 py-1 text-xs` (`px-4 py-2 text-sm` for the reply-time pill).

| Variant | Tokens |
|---|---|
| On a tint | `bg-ground text-ink` |
| Saving | `bg-[var(--accent-soft)] text-ink` |
| Reply time / order received | `bg-[var(--tint-2)] text-ink` |

**States** — `default`.

### 2.4 Form field and form

**Purpose** — every typed input: contact form, "ask the shop" form, checkout.

**Anatomy** — `[label] [control]`; the label sits above in `text-sm text-ink-soft`; the control is
`w-full appearance-none rounded-[var(--radius-sm)] border border-field bg-ground px-4 py-3 text-ink`
(a textarea has 2–4 rows). A hidden `website` field, `aria-hidden`, off-screen, is a spam trap and is not
drawn. Required fields use the `required` attribute; there is no asterisk.

**States** (the control)

| State | Visual change | Tokens |
|---|---|---|
| Default | As above | `--field-line`, `--ground` |

**States** (the form — what the design draws)

| State | Visual change | Tokens |
|---|---|---|
| Idle | Fields and the submit button | — |
| Sending | Button `disabled:opacity-60`, label "Sending…" | — |
| Sent | The form is replaced by one thank-you line (`text-lg`, `--ink`) | — |
| Failed | A `role="alert"` line under the fields (`text-sm`; `--danger` in the order and ask forms, `--ink` in the contact form) | `--danger` |
| Not connected | `text-sm text-ink-soft` line "This form is not connected yet — please try again later." above the button | `--ink-soft` |

**Accessibility** — every control is inside a `<label>`; `autoComplete` is set for name, phone, address;
`inputMode="numeric"` for height.

### 2.5 Select

**Purpose** — every choice from a list on the storefront (sort order, the frame size on the ask form).
Never a native `<select>`.

**Anatomy** — `[trigger: label + chevron]` and an open `[listbox]` under it. Trigger:
`rounded border border-field bg-surface px-3 py-2 text-sm`; list: `rounded border border-field bg-surface py-1 shadow-lg`, `max-h-64`, `z-20`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | The chosen label, or the placeholder in `--ink-soft` | `--surface`, `--field-line` |
| Open | The listbox under the trigger; the highlighted option is `bg-accent text-accent-ink` | `--accent`, `--accent-ink` |

**Accessibility** — `role="combobox"` with `aria-expanded`, `aria-controls`, `aria-haspopup="listbox"`;
the list is `role="listbox"` with `role="option"` and `aria-selected`. Arrows, Home, End, Enter, Space,
Escape, Tab and type-ahead work; an outside click or a scroll closes it.

### 2.6 FilterRail

**Purpose** — narrow the catalogue by frame size, wheel, brand and price. Choices live in the address
bar, so a filtered list is a link.

**Anatomy** — per group: `[title (xs, uppercase, ink-soft) + count of chosen] [checkbox list]`, a 1 px
rule under each group; then, only while filtering, `[Clear all]`. Rail width `lg:w-52`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Every group open; checkboxes unchecked (`accent-color: var(--accent)`) | `--field-line` |
| Filtering | A count in `--ink` beside each group title that has a choice; an underlined "Clear all" link appears | `--ink`, `--ink-soft` |
| Folded (below 1024 px) | A group can be closed by tapping its title (`<details>`); at `lg` the titles stop being toggles | — |

### 2.7 QuantityStepper

**Purpose** — how many of one bike, in the bag and the drawer.

**Anatomy** — `[−] [number] [+]` in a `rounded border border-line` group; buttons `p-2.5` with 16 px icons;
the number `min-w-8 text-center text-sm tabular-nums`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | `--ink` icons | `--line` |
| Disabled | `disabled:text-ink-soft` on the − at 1 and on the + at the stock limit | `--ink-soft` |

**Accessibility** — `role="group"` `aria-label="Quantity"`; the buttons are labelled "One fewer" / "One
more"; the number is `aria-live="polite"`.

### 2.8 Frame-size pill

**Purpose** — choose one frame size on a bike's page, with the rider height it fits underneath.

**Anatomy** — `[size (sm, 600)] [rider height (11 px, tabular)]` in a `min-w-[3.6rem] flex-1 rounded-[var(--radius-sm)] border px-2 py-2.5` pill; the group is a `fieldset` with `role="radiogroup"`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | `border-field text-ink` | `--field-line` |
| Hover | `border-ink` | `--ink` |
| Selected | `border-ink bg-accent text-accent-ink`; `aria-checked="true"` | `--accent` |
| Disabled (sold out) | `border-line text-ink-soft line-through cursor-not-allowed` | `--line`, `--ink-soft` |

### 2.9 Option card (checkout delivery and payment)

**Purpose** — choose how the bike reaches you and how you pay, as a large rounded card, never a box to type into.

**Anatomy** — `[icon] … [round check mark]`, then `[title (sm, 600)] [line (xs, ink-soft)] [extra (xs, 600)]`; `rounded border p-4`, three across from `sm`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | `border-field bg-ground`; an empty ring where the check goes | `--field-line` |
| Hover | `border-ink` | `--ink` |
| Selected | `border-ink bg-[var(--tint-1)]`; the ring fills navy with a check; `aria-checked="true"` | `--tint-1`, `--accent` |
| Disabled | `opacity-50 cursor-not-allowed` (cash on delivery when delivery is not "in town") | — |

### 2.10 ProductImage

**Purpose** — every bike photograph on the site. The frame decides the shape, never the file.

**Anatomy** — a `rounded` tile on `--product-ground` holding the picture; `4/3` for bikes. A `.jpg` /
`.webp` fills the tile edge to edge (`object-cover`); a `.png` with no background is shown whole with an 8 %
margin. Optionally a second picture for the pointer.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | The photograph in its tile | `--product-ground` |
| Loading | The white tile is seen until the picture arrives (lazy except in view on load) | `--product-ground` |
| No photograph | A flat block with the caption in `--product-ground-ink` | `--product-ground-ink` |
| Hover (level lively, inside a card) | The picture scales to 1.05 over 0.6 s; a card's second photograph fades in over 0.35 s | `--ease-out` |

**Accessibility** — `alt` names the bike (`Brand Model`), and the view when it is a close-up ("…, drivetrain
close-up"). The hover photograph is `alt=""` and decorative.

### 2.11 ProductCard (a bike in a grid)

**Purpose** — one bike in the shop grid, Home "more bikes" and the related row. Three to a row at `lg`,
one on a phone.

**Anatomy** — `[ProductImage 4/3, rounded, 1px --line edge] [heart, top right] [brand · riding type: xs uppercase, 0.14em, ink-soft] [name: base 600 | one-line fact: sm ink-soft] [price: base 700, right-aligned | old price: sm, struck through] [StockLine] [full-width outline button "Choose a size"]`.
A row takes turns in which of the bike's photographs it leads with (drivetrain, whole, fork), so two
white-sweep bikes never stand side by side.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | As above | `--line`, `--ink` |
| Hover | The name underlines; the picture swaps to the whole bike and scales; the outline button fills navy only when it is itself hovered | `--ease-out` |
| Out of stock | The StockLine reads "Out of stock" with a `--danger` mark; the button reads "Ask when it is back" | `--danger` |

### 2.12 TintCard (the featured bike)

**Purpose** — the three bikes the shop leads with on Home, and the bike of the month; one pale tint per card.
Not used in grids of many.

**Anatomy** — `[white pill: riding type] [name: brand light + model bold, 28/32 px] [one-line fact (sm, ink-soft)] [photograph window, rounded-[var(--radius-sm)], 4/3, edge to edge] [glass bar at the foot: price (sm 600) + round navy arrow button]`.
The card is `rounded` with `overflow-hidden`; the last word of the model is bound to the one before it
(non-breaking space) so "1" never stands alone.

| Variant | Tokens | When to use |
|---|---|---|
| Full | tint by `TINTS[i % 4]`, name `lg:text-[2rem]`, `p-5 lg:p-7` | The Home row of three; bike of the month |
| Compact | name `text-xl`, `p-5`, no fact line | A tinted card where space is short |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | As above | `--tint-1…4`, `--glass`, `--shadow-sm` |
| Hover (level lively, pointer device) | The card tilts under the pointer (`data-tilt`); the photograph swaps to the second picture; the arrow button `hover:opacity-90` | — |

**Accessibility** — the arrow button is labelled "Choose a size <model>"; the whole card links to the bike.

### 2.13 StockLine

**Purpose** — whether a bike can be had, in the owner's own words ("In stock" / "Out of stock"), with how
few are left when it is three or fewer. Always a word with its mark, never a colour alone.

**Anatomy** — `[icon 15 px] [word (500)] [· N left?]` in `text-xs` (`text-sm` on the bike's page).

**States**

| State | Visual change | Tokens |
|---|---|---|
| In stock | `done` icon and word in `--ink` | `--ink` |
| Few left (1–3) | As in stock, plus "· N left" in `--ink-soft` | `--ink-soft` |
| Out of stock | `alert` icon in `--danger`, word in `--ink-soft` | `--danger`, `--ink-soft` |

### 2.14 SaveButton (the heart)

**Purpose** — keep a bike for later; sits at the top right of a card tile and in the bike's tinted head.

**Anatomy** — a 32 px `rounded-full bg-ground` circle with a 16 px heart and `--shadow-sm`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Heart in `--ink-soft` | `--ink-soft`, `--shadow-sm` |
| Hover | Heart in `--ink` | `--ink` |
| Saved | Heart filled, in `--accent`; `aria-pressed="true"`, label "Remove … from saved" | `--accent` |

### 2.15 Rating and ReviewCard

**Purpose** — what a rider said and how many marks they gave. A rating is `value` filled stars of 5 — the
shape of the row is the information; five outlines would read as "nobody has rated this".

**Anatomy (Rating)** — five lucide stars, filled `--accent` for the value, outlined in `--line` for the
rest, 1.5 stroke, 15–16 px; `role="img"` with an `aria-label` "4 out of 5".

**Anatomy (ReviewCard)** — a `figure`: `[Rating … quote icon] [text (sm, relaxed)] [name (600) · bike | date]` in a `rounded border border-line bg-ground p-6`, with a hairline above the caption; the date is `formatDate(…, "medium")`.

**States** — `default` for both.

### 2.16 ProductGallery

**Purpose** — the bike's large picture and the others under it to switch to.

**Anatomy** — `[main ProductImage 4/3] [row of thumbnails: w-16 sm:w-20, rounded-sm border]`. On a phone the
main picture runs edge to edge. One picture is drawn alone, with no thumbnail strip.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | The first picture is shown; thumbnails at `opacity-70` with `border-line` | `--line` |
| Thumbnail hover | `opacity-100` | — |
| Thumbnail selected | `border-ink`, full opacity; `aria-pressed="true"` | `--ink` |
| Pointer on main picture (level lively) | The picture is looked into at twice its size (`data-lens`) | — |

### 2.17 SpecList, size-guide table and ProductTabs

**Purpose** — what the bike is made of (frame, fork, shock, motor, drivetrain, brakes, wheels, tyres,
weight, colour), what size to take, and the longer things — scanned, not read.

**Anatomy (SpecList)** — a `dl` in two columns (`minmax(7rem, auto)` and 1fr), `text-sm`, a hairline above each row; the label is `--ink-soft`, the value `--ink`. A row with no value is not drawn.

**Anatomy (size guide)** — a table: head `text-xs uppercase tracking-wider text-ink-soft`, rows `py-3` with a hairline; size in 600, rider height right-aligned and tabular, "In stock" / "Out of stock" in `--ink-soft`; a note under it in `text-xs`.

**Anatomy (ProductTabs)** — a tab list (About, Riding, Delivery & payment, Reviews) with a 1 px rule under it, then the panel (`max-w-[68ch]`, `pt-6`). A tab with nothing in it is left out; one tab alone is drawn as its content with no bar.

**States (a tab)**

| State | Visual change | Tokens |
|---|---|---|
| Default | `text-sm text-ink-soft`, a transparent 2 px underline | `--ink-soft` |
| Hover / focus-visible | `text-ink` | `--ink` |
| Active | `text-ink` with a 2 px `border-ink` underline | `--ink` |

**Accessibility** — Radix Tabs: `role="tablist"`, arrow keys move between tabs. The spec list is a real `dl`.

### 2.18 SiteHeader (with MenuToggle, MobileMenu and search)

**Purpose** — the way round the site: the mark and name at the left, the pages in the middle, search,
the owner's sign-in and the bag at the right.

**Anatomy** — one quiet row `h-[var(--header-h)]` between two `--line` hairlines that stop at the page's
measure: `[Logo] [nav (lg) | menu toggle (below lg)] [search button] [sign-in (lg)] [bag with count dot]`.
Nothing in it is coloured; the count dot is `bg-accent` with `text-[0.625rem]`.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Nav link default / hover | Plain; `hover:underline` with `underline-offset-4` | `--ink` |
| Nav link current | `font-semibold` | — |
| Icon button hover | `bg-wash` circle | `--wash` |
| Bag with items | A 16 px navy dot with the count at the icon's top right | `--accent`, `--accent-ink` |
| Search open (lg) | A field row under the header: search icon, 18 px input, "Search" button, a hairline below; the search icon turns into a close icon | `--line` |
| Search open (below lg) | A solid panel over the screen under the header with a 30 px light title, a full-width field and a button | `--ground`, `--field-line` |
| Menu open (below lg) | The toggle turns into an X; a solid panel with the pages at `text-3xl font-light`, "Bag" and the sign-in link | `--ground` |

**Accessibility** — `nav aria-label="Main"`; the search button and the toggle carry `aria-expanded`;
the mobile panels are `role="dialog" aria-modal="true"`, lock the page scroll and close on Escape; the
search form is `role="search"`.

### 2.19 SiteFooter

**Purpose** — the site's one surface of navy: a large rounded card at the page's measure that closes
every page, as the same ink presses every button.

**Anatomy** — `[mark + name, one line] | [Riding types] | [Pages] | [Reach us: phone, email, address, hours, message, shop sign-in]`, then a rule at `accent-ink/20`, the independence line and the photo credits at `text-xs`. A way to reach the shop is drawn only where the owner has given it; the message link always is.
Text is `--accent-ink` at 75 %, full on hover; titles are 600.

**States** — `default` (links have the hover above).

### 2.20 CartDrawer (the bag as a panel)

**Purpose** — see what is in the bag without leaving the page; the way to check out.

**Anatomy** — a panel from the right edge, `max-w-md`, `bg-ground`, over a `bg-black/40` scrim: `[title | close] [lines: ProductImage 80 px, name, size, line total, QuantityStepper, remove] [total | Check out (a primary block link)]`. Radix `Dialog` carries the focus trap, scroll lock and Escape.

**States**

| State | Visual change | Tokens |
|---|---|---|
| Empty | A centred `--ink-soft` line, no footer | `--ink-soft` |
| Filled | Lines divided by hairlines; the total in `text-lg` | `--line` |

### 2.21 ContactDock

**Purpose** — call and Zalo within a thumb's reach on a phone: two 56 px round buttons fixed to the lower
right corner below 1024 px. Each shows only when its number is in `content.json`; with neither, nothing shows.

**Anatomy** — `[Zalo: bg-surface, ring-1 ring-line, the word "Zalo"] [Call: bg-accent, phone icon]`, `shadow-lg`, `z-50`, with the page body padded by `--dock-space` so the foot of the page is never covered.

**States** — `default`.

### 2.22 Bands and panels

Surfaces that carry colour or structure on a page. Each is drawn once and has one state, `default`.

| Panel | Anatomy | Tokens |
|---|---|---|
| Bleed band (catalogue head, bag head) | A `--surface` band to both screen edges with the heading at the page's edge and a landscape photograph off the right edge (a strip on top on a phone) | `--surface` |
| Promises band (Home) | Three centred columns: a 48 px white round mark, a 600 line, a `text-xs` line | `--tint-4` |
| Summary card (bag) | Subtotal and delivery rows, the glass total row, the Check out button, then the promise list | `--tint-1`, `--glass` |
| Order-button panel (checkout) | What happens next, the error / not-connected lines, then Place order | `--tint-2` |
| Help band (under the catalogue) | A line, a sentence and a button, run off the right edge | `--tint-3` |
| Ask band (bike page) | Heading left, the ask form right, run off the right edge | `--surface` |
| "Where to start" (Home) | A grey band: heading, then a row per riding type (name 600, who it is for, terrain, round arrow) between hairlines; a photograph at the right | `--surface`, `--line` |
| Promise list (bike page, bag) | A bordered box: delivery, warranty, reply time, each with an icon | `--line` |
| Receipt steps (checkout done) | Four steps on navy; the current one's mark is butter | `--accent`, `--tint-2` |
| Departments row (Home) | The shop's riding types as grey letterspaced wordmarks between two hairlines, each a link (`--ink-soft`, hover `--ink`) | `--line` |
| Proof block (Home hero) | A short rule, three overlapping 48 px round bike photographs (`ring-2` in `--ground`), the in-stock count counted from the records and a line | `--ground`, `--line` |

### 2.23 Empty, loading and missing states

| Where | What the design draws | Tokens |
|---|---|---|
| Shop, no bike matches | A `--tint-1` panel: a `text-xl` headline, a `text-sm` line saying what to change, a primary "Clear" button | `--tint-1` |
| Shop, loading | A single `text-sm text-ink-soft` line | `--ink-soft` |
| Bag, empty | A `--tint-1` panel: `text-2xl` headline, a line, the riding-type chips and a primary "Browse" button | `--tint-1` |
| Bike not found | A `text-[length:var(--size-h2)]` headline and a primary button back to the shop | — |
| Nothing ordered yet (receipt) | The navy bleed band: a heading, one line and an inverted button | `--accent` |

Every empty state says what is missing and gives the next action; none is a blank area.

### 2.24 Grid

**Purpose** — a row of like things (bikes, reviews, promises) that never leaves a hole in its last row.

**Anatomy** — a wrapping flex row, `gap` = `--space-6`, last row centred; the column count comes from the
number of cards (`columnsFor`): 1 or 2 on a phone (`phone`), at most 3 from `sm`, at most `max` from `lg`.
One card size and one picture ratio per grid.

**States** — `default`.

### 2.25 Breadcrumb

**Purpose** — where a page sits: Home › All bikes › Trail › Model.

**Anatomy** — `nav aria-label="Breadcrumb"`, `text-sm text-ink-soft`, 14 px chevrons between; the current
page is `--ink` and not a link; Home is a house icon with a label.

**States** — Default; Hover (`text-ink`).

### 2.26 Logo

**Purpose** — the mark and the name. The mark is a saddle between two peaks — a mountain pass — drawn in
`currentColor` (`LogoMark`, `public/favicon.svg`); beside it the name in the display face at `text-lg`
with `tracking-tight`. An owner's own logo (`site.logo`) is used as given and never redrawn.

**States** — `default`.

### 2.27 Management area kit

`/admin` uses the shadcn/ui kit in `src/components/ui` (button, input, select, table, dialog, sheet,
sidebar, tabs, badge, switch…) inside `DashboardLayout`, in the fixed palette of 1.1. Its controls draw
**focus-visible** as a 3 px `ring-accent/50` with `border-accent`, **disabled** as `opacity-50`, and
**invalid** as `border-danger` with a `ring-danger/20`. It is a tool, not the shop's look, and stories for
the owner's site do not restyle it.

## 3. Content and formatting

- **Words live in `content.json`.** A component renders `<T k="…"/>` or `useContent("…")`; a story never
  hard-codes a visible string, and the owner edits every word in the Studio. The shop's name is **Đèo Bikes**,
  written exactly so, with its diacritics.
- **Voice** — a shop that rides what it sells: plain, direct, warm, second person ("Choose a size",
  "Ask the shop to call", "Take one filter off — or tell us what you need and we will order it in"). Short
  sentences, an em dash for an aside, no exclamation marks, no hype. The site is written in English.
- **Locale** — `site.locale` is `vi-VN`. Every price and date goes through `src/lib/format.ts`, never
  `toLocaleString`, `Intl` or `${price} ₫` in a component:
  - Price: `formatPrice(n)` → `1.200.000 ₫` (currency VND, from the price fields in `schema.json`).
  - Date: `formatDate(d)` → `16/10/2026`; `"medium"` → `16 Oct 2026` (reviews); `"long"` →
    `Friday, 16 October 2026`. Time: `formatTime` → `14:30`; both: `formatDateTime` (the receipt).
  - Figures (prices, counts, rider heights, quantities) use `tabular-nums`.
- **Capitalisation** — sentence case for buttons, headings, labels and chips ("Place order", "Bike of the
  month", "Your name"); proper nouns keep their capitals (Giant, Liv, Zalo, VietQR, E-MTB). Uppercase is
  made by CSS (`uppercase` on eyebrows and wordmarks), never typed.
- **Stock words** — "In stock", "Out of stock", and "· N left" when three or fewer remain; the owner's
  words, always with a mark, never a colour alone.
- **Empty and error wording** — say what is missing, then what to do: a headline in the form "No bike
  matches all of that." / "Your bag is empty.", one line of what to change or where to start, and a button.
  A failed send shows the message the sender returned in a `role="alert"` line; a form with no sender
  connected says so before the button.
- **Delivery promises** are one record (`src/data/shop.json`) shown by `DeliveryTerms` on Home, the bike's
  page, the bag and the checkout, so no page can promise a different thing.

## 4. Known deviations

Places where the approved design does not follow its own rules or the anti-patterns in
`references/ai-defaults.md`. Recorded, not fixed: the owner approved the design as it stands.

| Where | Deviation | Why it stands | Follow-up |
|---|---|---|---|
| `--field-line` on `--ground` (every input, Select, size pill, option card, checkbox) | `rgba(25,26,46,0.45)` flattens to about `#9898A1`, **2.9:1** against white — under the 3:1 a control's edge needs, and under `theme.css`'s own comment ("at least 3:1 against `--ground`") | The palette is the owner's picture of the template; the shortfall is small | Raise the alpha to about `0.50` (≈ 3.3:1) in `theme.css` — a scaffold edit for TL, not a story |
| Whole storefront | **No focus style of its own and no focus token.** Links and buttons rely on the browser's default ring. `outline-none` with no replacement: the header's search field and the open tab panels; Tabs triggers only change text colour on `focus-visible`; the CartDrawer panel is `outline-none` (Radix moves focus inside it) | Nothing was drawn for it; the default ring is still shown on every plain link and button | Add a `--focus` token and a ring recipe; never remove the default ring without one |
| Header icon buttons (40×40), SaveButton (32×32), bag remove (40×40), QuantityStepper buttons (~36 px), chips (~36 px high) | Below the 44×44 touch minimum on a site read mostly on phones | Sized to the template's compact pills | Raise the hit area (padding or a `min-h-11`) in a later story without changing the drawn size |
| `Rating` (unfilled stars), QuantityStepper edge, promise and contact boxes | Drawn in `--line`, ~1.3:1 on white | A hairline is decorative; the star count is also given as an `aria-label` and the number beside it | None needed for the boxes; consider `--field-line` for the stepper edge |
| Buttons, chips, pills (`rounded-full`) and cards (`--radius` 24 px) | AI default 3 — maximum rounding: every control is a pill and every card is 24 px | It is the template's own shape; the scale still has three steps by element type (24 px surfaces, 14 px fields, pill controls) | None; do not add a fourth radius |
| Home and product page | AI default 4 — `pt-24` (96 px) before every major section | Marketing sections get generous padding; dense lists and tables use `py-3`–`py-5` | None |
| `CartDrawer` scrim `bg-black/40`; floating layers `shadow-lg` (Select list, ContactDock) and `shadow-xl` (CartDrawer); the management kit's `text-white` and `bg-black/50` | Values from outside the theme: Tailwind's black and its default shadows are not tokens | They are genuinely floating layers, where a strong shadow is right (AI default 5 is about resting cards, which have none) | Add `--scrim` and `--shadow-lg` to `theme.css` (TL) and point the layers at them |
| Type | Arbitrary sizes off the ramp: `text-[0.625rem]` (10 px), `text-[0.6875rem]` (11 px), `text-[0.9375rem]` (15 px), `text-[1.75rem]`/`lg:text-[2rem]`; arbitrary tracking `0.14em`, `0.2em`, `-0.01em`; `leading-[1.08]`–`[1.12]`. 10 px and 11 px are below 12 px | The tinted card and the size pill need sizes the ramp has no step for | Add ramp steps if more screens need them; do not copy the arbitrary values into new components |
| `ContactForm` | Its `role="alert"` failure line is `text-ink`; the order and ask forms use `text-danger` | Not caught in review | Use `--danger` for a failed send everywhere |
| `--shadow-md`, `--duration-fast`, `--duration-base` | Defined in `theme.css` and used by no storefront element; the real transitions are 0.6 s (zoom), 0.35 s (swap) and Tailwind's default colour transition | The scaffold ships the scale whether or not a design uses it | New transitions should use the tokens |
| Pointer devices | The second photograph, the 1.05 zoom, the tilt, the magnetic pill and the gallery lens answer only a hover; a phone gets none of them | None carries information a phone needs — every picture is reachable from the gallery thumbnails | None |
| `content.json` `contact.phone`, `contact.zalo`, `contact.email`, `contact.address`, `contact.hours` | **All empty.** So the ContactDock, the footer's reach rows, the contact page's detail box and the map draw nothing; only the message form (always drawn) is a way to reach the shop | The owner has not given them yet; the design draws them only where given | Owner enters them in the Studio — the brief asks for "a way to call or message" |
| `src/components/SectionHead.tsx`, `VariantPicker.tsx` | Exist in the scaffold; no page draws them, so they are **not part of this design system** | The bike's page uses its own size pills | A story that needs one must have it designed first |

**Checked and clear.** No gradient anywhere. No indigo or violet default: the ink is navy `#191A2E` and the
lilac `#EFE2FE` is a pale card tint, not an accent. No emoji — one icon set, `lucide-react`, at one 1.75
stroke through the `Icon` map (`shipping` is a truck on every page). No lorem ipsum. Text is never laid over
a photograph without a scrim (the hero headline sits beside its photograph; the price sits on a glass bar).
Every empty list says what is missing and what to do. Icon-only controls carry an `aria-label`.

## 5. Change log

| Date | Change | Design PR |
|---|---|---|
| 2026-11-18 | First written: extracted from the approved design (`theme.css`, `index.css`, `content.json`, `App.tsx`, `components/`) | #1 |
