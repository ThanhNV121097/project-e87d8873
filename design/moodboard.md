# Đèo Bikes — how a mountain-bike shop sits in Pastel

The owner chose the template **Pastel** by looking at it (`design/template.jpg`): a white page,
one near-black navy for type and everything pressed, three tall cards in pink, butter and lilac
with a glass price bar across their foot, and headings whose words change weight mid-line. That
look is settled. This file says how **Đèo Bikes** — a shop selling mountain bikes for passes,
trails and rough ground, to first-timers and long-time riders, mostly read on a phone — fills it.

**The name.** The owner gave none. *Đèo* is the word the brief itself uses for the roads these
bikes are bought for (đường đèo, a mountain pass). Short, local, says what the shop is for.
The mark (`LogoMark.tsx`, `public/favicon.svg`) is a saddle between two peaks — the shape of a pass.

## What the shop puts in each device

| Template device | What it holds here | Component |
|---|---|---|
| TINTED PRODUCT CARD | The three bikes the shop leads with (one hardtail, one trail bike, one e-MTB), each on pink / butter / lilac: the riding type as the white pill, the model name large, the maker's studio photograph filling the lower window of the card, the glass bar with the price and the round dark bag button | `TintCard.tsx` |
| MIXED-WEIGHT HEADING | "Built for / Passes. Trails / & rough ground" and every section head ("Bike of / the month", "Not sure / where to start?") | `Weights.tsx`, hero in `Home.tsx` |
| INLINE PILL | "See all bikes" on the hero heading's last line | `Home.tsx` (`data-testid="hero-cta"`) |
| PROOF BLOCK | A short rule, three of the shop's own bikes as overlapping rounds, the count of bikes in stock (counted from the records, so never a claim) and two lines | `Home.tsx` |
| PARTNER MARK ROW | No partner logos exist (and no real company's marks are borrowed): the shop's own departments — HARDTAIL · CROSS-COUNTRY · TRAIL · E-MTB · WORKSHOP — as grey letterspaced wordmarks between two hairlines, each a link | `Home.tsx` |
| FEATURE ROW | The peach band: delivered assembled · two-year frame warranty · a person answers within two hours — each a round mark, a bold label, a line. Their numbers live in `src/data/shop.json`, marked assumed | `Home.tsx` |

## What I moved from the template, and why

- **Hues and faces: nothing moved.** The owner picked this palette by looking; navy `#191A2E`,
  the page `#FFFFFF`, the band `#F6F6F6`, tints `#FFEBED` / `#F6FFB0` / `#EFE2FE` / `#FFEFE2`, Inter
  Light against Inter Bold. A bike is a hard, mechanical object; the soft tints read as a calm,
  trustworthy shop around it, and the energy comes from the photographs of riding.
- **`--product-ground` is white, not grey.** Every bike picture is its maker's studio photograph on
  a white sweep. A grey tile under a white photograph is a white box; a white tile with a hairline
  edge lets photograph and tile be one surface. That is the template's own rule (the tint and the
  picture's ground agree) applied to the grid card.
- **The tinted card's photograph sits in the lower window, not across the card.** The template's
  photographs were shot on the card's colour. Free stock has no cycling shot on a flat pastel
  ground — `fetch_photos` measured eight candidates for each of pink, butter and lilac and found
  none — so, as the tool directs, the colour is carried by the card itself (the header with the
  pill and the name) and the bike's own white-sweep photograph fills the window below, edge to edge.
- **Bikes are landscape objects.** Every bike frame on the site is 4:3, and the makers' image
  server pads each one to exactly that, so no wheel is ever cropped.

## What the template has that this shop cannot make

- **Partner logos** → the shop's own departments as wordmarks (above).
- **"15K++ positive reviews"** → the count of bikes in the shop, read from the records.
- **A styled pastel shoot** → see the tinted card above. The month section and the gallery use
  real riding photographs (stock, credited) — the template's own month cards are outdoor shots.
- **Faces in the proof rounds** → the shop's own bikes; never strangers' faces.

## Where the records and pictures come from

- **Bikes (14):** real 2025–2027 models from **Giant** and **Liv** (Giant's women's brand). Names,
  prices, frame, fork, drivetrain, brakes, sizes and rider heights were read from the makers'
  pages: https://www.giant-bicycles.com/au/talon-2 ,
  https://www.giant-bicycles.com/au/trance-x-29-2 ,
  https://www.giant-bicycles.com/au/bikes/mountain-bikes/trail ,
  https://www.giant-bicycles.com/au/bikes/mountain-bikes/cross-country ,
  https://www.giant-bicycles.com/au/bikes/mountain-bikes/electric-mountain ,
  https://www.liv-cycling.com/au/bikes/mountain-bikes — and one real Vietnamese listing with its
  own VND price: https://xedap.vn/shop/xe-dap-dia-hinh-mtb-giant-talon-29-3-phanh-dia-banh-29-inches-2025 .
  Australian prices are converted to sample VND and keep the page's figure as `price_listed`.
  Weights (Giant does not publish them), ratings, review counts, stock and the shop's own story
  are invented and named in each record's `_invented`.
- **Bike pictures (40):** each bike's own studio photograph plus two close-ups (drivetrain;
  cockpit and fork) cropped from that photograph by the maker's image server — borrowed, under
  `public/images/ref/`, to be replaced by the shop's own.
- **Scenes (10):** Pixabay, credited in `src/credits.json` and shown by `<PhotoCredits/>`.

## Motion — LIVELY (the owner's choice)

From `src/motion`, nothing written by hand: the hero heading's words rise (`SplitHeading`), the
hero pill is magnetic (`data-magnetic`), the bike count counts up (`data-count`), the three tinted
cards tilt under the pointer (`data-tilt`), their photographs drift a few percent against the
scroll (`data-parallax`), grid cards swap to the drivetrain close-up under the pointer
(`ProductImage hover`), and a bike's main picture is looked into under the pointer
(`ProductGallery`). Native scrolling; nothing pinned or scrubbed.

## Unlocks

None. The template already sets the weight and scale (Inter 300 against 700, h1 56px, h2 34px);
the house-style ceilings are untouched.

## THE PAGES NOBODY DREW

`design/template.jpg` shows the home page. Every page below is designed here.

### /shop
- **Who arrives:** someone who knows roughly what riding they do and a budget, narrowing fourteen bikes to two or three.
- **First screen (1440×790):** the mixed-weight heading with the live count at its right; a row of type chips; the filter rail (type, frame size, wheel, brand, price) beside three large bike cards to a row with their price, wheel, travel and stock line.
- **The look:** the MIXED-WEIGHT HEADING ("All the / bikes"); the template's filter-chip row (the "Favorite Picks" pills) as riding-type chips with the INLINE PILL's dark fill for the chosen one; grid cards on white tiles with the heart, name and price on one line and the outlined pill, as the template's picks.
- **From what I looked at:** https://www.giant-bicycles.com/au/bikes/mountain-bikes/trail — filters a rider uses (suspension, wheel size, frame material) and a count of results; https://xedap.vn/shop/xe-dap/xe-dap-dia-hinh — brand and a price range as filters, a struck-through price beside the sale price. Taken: the filter questions and the count; three cards a row rather than four, because fourteen bikes are few and each needs its figures.

### /shop/:id
- **Who arrives:** a rider deciding on one bike and one size — or deciding to ask the shop first.
- **First screen:** the bike's picture large (about 60% of the width, 4:3) with its two close-ups under it; at the right the riding-type pill, name, price set large, the stock line, the frame-size pills each with the rider height it fits, quantity, "Add to bag" and "Ask about this bike".
- **The look:** the tinted header of the TINTED PRODUCT CARD carries the name and pill above the picture; the INLINE PILL is "Add to bag"; the FEATURE ROW returns as the delivery / warranty / reply line under the button.
- **From what I looked at:** https://baymard.com/mcommerce-usability/benchmark/mobile-page-types/product-page — the picture first, then price and the action before any description, specifications as a scannable table; https://www.giant-bicycles.com/au/talon-2 — a size guide by rider height (S 163–172 cm…), and the full spec list (frame, fork, drivetrain, brakes, tyres, weight); https://xedap.vn/shop/xe-dap-dia-hinh-mtb-giant-talon-29-3-phanh-dia-banh-29-inches-2025 — the "in stock at the shop" line and the hotline beside the buy button.

### /bag
- **Who arrives:** someone checking what they chose and what it will cost before giving details.
- **First screen:** the heading and the lines — each the bike's picture, name, chosen size, quantity stepper, line total right-aligned — beside a summary card with subtotal, delivery, total and "Check out".
- **The look:** MIXED-WEIGHT HEADING ("Your / bag"); the summary card is a tinted (pink) panel with the glass price bar idea as its total row; the FEATURE ROW promises under it.
- **From what I looked at:** https://baymard.com/checkout-usability/benchmark/step-type/cart — the cart's lines with picture and editable quantity, the total saying what it includes, and an empty cart that points back to the goods.

### /checkout
- **Who arrives:** a buyer who has chosen and needs to say where it goes and how to pay.
- **First screen:** the form (name, phone, delivery or collection, address, payment as marked options) at the left; the order — pictures, lines, delivery charge, total — in view at the right.
- **The look:** MIXED-WEIGHT HEADING; payment and delivery choices as large rounded option cards with a tinted fill when chosen (the TINTED card in small); the dark INLINE PILL "Place order"; a line saying what happens after (the shop calls to confirm).
- **From what I looked at:** https://baymard.com/checkout-usability/benchmark/step-type/cart and https://giant.vn/shop/xe-dap-dia-hinh-mtb-giant-talon-29-1-2022/ — delivery and collection from a showroom as options, and the note that a showroom confirms stock before the sale is final.

### /checkout/done
- **Who arrives:** a buyer who just placed an order and wants to know it worked and what happens next.
- **First screen:** the answer, largest ("Order received"), with the order number; the steps as one line (Received → Confirmed by phone → Assembled → Delivered) with where it is now; then what was ordered with its picture, where it goes, and how it is paid.
- **The look:** MIXED-WEIGHT HEADING; a tinted (butter) card holding the steps; the FEATURE ROW's round marks for each step.
- **From what I looked at:** https://baymard.com/checkout-usability/benchmark/step-type/receipt — the confirmation first, then the order's summary, delivery details and next steps.

### /gallery
- **Who arrives:** a rider who wants to see what this kind of riding looks like before buying a bike for it.
- **First screen:** the heading at the left with a line, and the first photograph wide beside it; then the rest large, two to a row, one landscape frame each, a caption under each.
- **The look:** MIXED-WEIGHT HEADING ("Where / these bikes / go"); the PARTNER MARK ROW's hairline-and-wordmark rule as the caption line under each photograph; the INLINE PILL ("See the bikes") on the closing line.
- **From what I looked at:** https://www.giant-bicycles.com/au/bikes/mountain-bikes/trail — a full-width riding photograph opening the list; https://dribbble.com/search/contact-page — the "About us" shots that set photographs large in an uneven two-column rhythm.

### /about
- **Who arrives:** someone deciding whether to trust a shop with a 15–300 million đồng purchase.
- **First screen:** the story heading and three lines at the left, the workshop photograph large at the right, and the shop's numbers (years, bikes built, rating) set large under the words.
- **The look:** MIXED-WEIGHT HEADING; the PROOF BLOCK (rule, count, two lines) carries the numbers; the TINTED cards carry the team (names and roles, no faces) and the reviews; the workshop price list as a right-aligned table.
- **From what I looked at:** https://originbike.com.vn/ — a Vietnamese bike shop's customer reviews with first names and star ratings, and its FAQ (genuine parts, choosing a size, warranty); https://dribbble.com/search/contact-page — the "About us" shot pairing a large photograph with a row of numbers.

### /contact
- **Who arrives:** someone with a question about a bike, a size or an order.
- **First screen:** the heading, the message form (name, phone or email, the bike, message) at the left; at the right a tall photograph and the ways to reach the shop as a list with icons — shown only where the owner has given them.
- **The look:** MIXED-WEIGHT HEADING; the dark INLINE PILL "Send message"; a FEATURE ROW of three reasons people write (size, test ride, workshop).
- **From what I looked at:** https://dribbble.com/search/contact-page — the form beside the details panel, the details as a short list with icons rather than a paragraph.

### /admin/sign-in
- **Who arrives:** the owner, to manage bikes and stock. It is the door to the management area, which keeps the scaffold's own neutral look (`[data-dashboard]`), not this site's.
- **First screen:** the sign-in form with the shop's name.
- **The look:** the shop's mark and name; the dark pill button.
- **From what I looked at:** https://dribbble.com/search/contact-page — the single-form screen with one clear action.

## Crit notes (what changed after watching the build)

- **Round 1 (preview at 1440 and 390, then the walk):** every page stopped at the same margin. Each page now
  lets ONE thing run off the screen's edge — the Bike-of-the-month photograph off the left on the home page,
  the sizing band off the right on `/shop`, the "Ask about this bike" band on a bike's page, the photographs
  on `/gallery`, `/about` and `/contact`, the steps band on `/checkout/done`. The home page's first screen
  keeps the template's arrangement, which has no bleed.
- On a phone the Bike-of-the-month card was squeezed beside its photograph and its price bar clipped: the
  photograph now runs edge to edge above a full-width card. The department wordmarks wrap left-aligned
  at a smaller size instead of being pushed to both edges.
- The footer's photo credits were 2.9:1 on navy; now they're above 7:1.
- The tinted cards lost 24px of air under the name so more of each bike shows in the first screen.
- **Round 2:** the management area's overview figure for stock value was too wide at 1180px; it is now the
  count of bikes out of stock, and "Running low" is a searchable, filterable, paged table.

## COMMITMENTS

1. At 1440×790 the first screen has no photograph above the cards: a three-line heading whose words change between Inter Light and Inter Bold ("Built for / Passes. Trails / & rough ground") with the dark "See all bikes" pill on its last line, and the tops of the three tinted cards in view.
2. At the heading's right: a short rule, three overlapping round crops of the shop's own bikes, the number of bikes in bold (counted from the records) and two lines of small text.
3. Under the heading, three tall rounded cards tinted pink, butter and lilac: a white pill with the riding type, the model name large, the maker's photograph filling the lower window, and a glass bar across the foot with the price at the left and a round dark bag button at the right.
4. The departments — HARDTAIL · CROSS-COUNTRY · TRAIL · E-MTB · WORKSHOP — stand as grey letterspaced wordmarks in one row between two hairlines, each a link.
5. A peach rounded band carries three promises side by side, each a round white mark, a bold label and one line.
6. Every bike card in a grid is a white 4:3 tile with a hairline edge holding the maker's photograph, a heart at its top right, the name and the price on one line, the stock status in the owner's words ("In stock" / "Out of stock") with its icon, and an outlined pill.
7. On a bike's own page, frame sizes are pills that each show the rider height they fit (M · 169–179 cm), and the specification is a two-column table: frame, fork, shock, drivetrain, brakes, wheels, tyres, weight.
8. Lively motion a reviewer can watch: the hero heading's words rise one after another, the hero pill leans to the pointer, the bike count counts up, the three tinted cards tilt under the pointer, grid cards show the drivetrain close-up under the pointer, and the bike page's main picture is looked into under the pointer.
9. One thing per page runs off the screen's edge: the Bike-of-the-month photograph off the left edge of the home page, the lilac sizing band off the right edge of `/shop`, the grey "Ask about this bike" band off the right edge of a bike's page, and the opening photograph off the right edge of `/gallery`, `/about` and `/contact`.
