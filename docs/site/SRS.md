# SRS — Site (the storefront)

Module: `site`
Design: [View the approved design](http://localhost:9400)
Design system: `design/design-system.md`

> One file per module. This one covers the storefront a customer visits: every page between the
> header and the footer, as the approved design draws it. The management area is drawn by the same
> design and is specified in `docs/admin/SRS.md`.

## 1. Purpose

Đèo Bikes sells mountain bikes to riders who want to ride passes, trails and rough ground, from a first
hardtail to a carbon race bike. The storefront lets a visitor — mostly on a phone — see the bikes the
shop leads with, browse and narrow every bike, open one to read its price, photographs, specification
and stock, put it in a bag and check out, or leave details so the shop calls back; and it tells them who
the shop is and how to reach it. Without it there is no shop.

## 2. Actors

| Actor | Who they are | What they may do in this module |
|---|---|---|
| Visitor | Anyone who opens the site, signed in or not; most are on a phone | Everything on the storefront: browse, filter, search, open a bike, keep a bike with the heart, fill the bag, check out, send a message |
| Shop owner | The person who runs Đèo Bikes | Reads what visitors send in Agent Team; changes the site's words and colours in the Studio; keeps bikes true in the management area (admin module). Has no extra storefront permission |

The storefront has no accounts and no roles: a visitor is never asked to sign in.

## 3. Scope

**In scope** — the functions specified below, by their plan titles:

- The site as designed
- Bike page waits for its record

**Out of scope**

- The management area (sign-in, overview, bikes, riding types) — belongs to the `admin` module.
- Where the records live and how lists are paged, searched and sorted — belongs to the `data` module. This
  module states what a page shows given the records; it does not care whether they come from a file or an API.
- What the server does with a placed order (prices, stock hold, order number) — belongs to the `orders`
  module. This module covers only what the checkout page and the confirmation page draw.
- Card payment on the site — deliberately not built: the shop takes bank transfer, cash on delivery and
  card or cash at the shop, and confirms by phone before anything is charged.
- A page listing the bikes a visitor has kept with the heart, customer accounts, a form for customers to
  write reviews, emailing the customer — none is drawn in the design; not built.
- A dark palette or a theme switch — the design has one mode.

## 4. Functional requirements

### 4.1 The site as designed

**Requirement SITE-001 — The frame: header, menu, search, footer, contact dock**

*As a* visitor, *I want* the same way round the shop on every page, *so that* I can reach any page, search, my bag and the shop's contact details from anywhere.

Behaviour:

1. Every storefront page is drawn between the header and the footer. The header is one row between two hairlines and is not sticky.
2. From 1024 px the header shows, left to right: the logo (the mountain-pass mark and "Đèo Bikes", linking to Home), the links "Bikes", "Gallery", "About", "Contact", a search button, a sign-in icon labelled "Shop sign-in" that opens the management area, and a bag button labelled "Bag". The link of the page the visitor is on is semibold.
3. The bag button carries a small dot with the number of bikes in the bag (the sum of quantities) when the bag has any, and no dot when it is empty. The button opens the bag drawer (SITE-007).
4. Below 1024 px the links and the sign-in icon are replaced by a menu toggle. Opening it shows a full panel with the four links at the large light size, then "Bag" and "Shop sign-in"; the toggle becomes a close mark; Escape or the toggle closes it; the page behind does not scroll while it is open.
5. The search button opens a field under the header on 1024 px and up, and a full panel titled "Search the shop" below 1024 px. Both have a field whose hint reads "Bikes, workshop jobs, questions — Stance, brake bleed, size" and a "Search" button. Submitting goes to the search page with the typed words; submitting an empty field goes to the search page with nothing typed. While the field is open the search icon is a close mark.
6. The footer is one large rounded card in the shop's navy at the page's measure. It holds: the mark and "Đèo Bikes" with the line "Mountain bikes for passes, trails and rough ground — chosen, built and set up by people who ride them."; a "Riding types" column listing each riding type, each linking to the shop narrowed to it; a "Pages" column listing the four header links and "Your bag"; a "Reach us" column; then a rule and, under it, the independence line and the photo credits.
7. "Reach us" lists the shop's phone, email, address and opening hours each only when the shop has given it, then always "Send us a message" (to the contact page) and "Shop sign-in" (to the management area). A phone or email is a `tel:` / `mailto:` link.
8. Below 1024 px, while the shop has given a phone number or a Zalo number, two round 56 px buttons are fixed to the lower right corner: "Zalo" (label "Message us on Zalo", opens the shop's Zalo chat in a new tab) and a phone (label "Call us", dials the number). A button is drawn only for a number the shop has given; with neither nothing is drawn. From 1024 px the buttons are not drawn. The foot of every page is padded so the buttons never cover content.
9. Every page sets the browser tab title to "<page title> — Đèo Bikes" (the home page: "Mountain bikes for passes and trails — Đèo Bikes"), and its description for search results and shared links, from the owner's words.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A 1280 px wide window | Any storefront page opens | The header shows the logo, then "Bikes", "Gallery", "About", "Contact", then the search button, the "Shop sign-in" icon and the "Bag" button |
| AC-2 | The visitor is on /gallery | The header is read | "Gallery" is semibold and the other three links are not |
| AC-3 | The bag holds 2 of one bike and 1 of another | The header is read | The bag button's dot shows 3 |
| AC-4 | The bag is empty | The header is read | The bag button has no dot |
| AC-5 | A 390 px wide window | A page opens | No nav links and no sign-in icon are drawn; a menu toggle is drawn |
| AC-6 | A 390 px wide window | The menu toggle is pressed | A panel lists Bikes, Gallery, About, Contact, Bag and Shop sign-in; the toggle shows a close mark |
| AC-7 | The menu panel is open | Escape is pressed | The panel closes |
| AC-8 | A 1280 px wide window | The search button is pressed | A field with the hint "Bikes, workshop jobs, questions — Stance, brake bleed, size" and a "Search" button appear under the header, and the search icon becomes a close mark |
| AC-9 | The search field is open and "bleed" is typed | Enter is pressed | The visitor lands on the search page with "bleed" searched |
| AC-10 | The search field is open and empty | The "Search" button is pressed | The visitor lands on the search page with nothing typed |
| AC-11 | A 390 px wide window | The search button is pressed | A full panel titled "Search the shop" shows a full-width field and a "Search" button |
| AC-12 | Any storefront page | The footer is read | It shows the mark and name, the shop's one-line description, the four riding types, "Bikes", "Gallery", "About", "Contact" and "Your bag", "Send us a message", "Shop sign-in", the independence line and the photo credits |
| AC-13 | The shop has given no phone, email, address or hours (as shipped) | The footer's "Reach us" column is read | Only "Send us a message" and "Shop sign-in" are listed |
| AC-14 | The shop has entered a phone number | The footer's "Reach us" column is read | A phone row with that number, a `tel:` link, is listed |
| AC-15 | A 390 px wide window and the shop has entered a phone number | Any page is read | A round phone button labelled "Call us" is fixed to the lower right and the foot of the page is not covered |
| AC-16 | A 390 px wide window and the shop has entered a Zalo number | Any page is read | A round "Zalo" button labelled "Message us on Zalo" is drawn beside the phone button, if any |
| AC-17 | The shop has entered neither a phone nor a Zalo number (as shipped) | Any page is read | No contact buttons are drawn |
| AC-18 | A 1280 px wide window | Any page is read | No contact buttons are drawn |
| AC-19 | Any page below the footer card | The footer's photo credits line is read | The credits for the stock photographs are shown under the label "Photos" |
| AC-20 | The visitor opens /shop | The tab is read | Its title is "Mountain bikes — Đèo Bikes" |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Contact details not given | The shop's phone, Zalo, email, address and hours are empty, as shipped | Their rows and buttons are not drawn; the message form is the way to reach the shop (SITE-012). See §7, open question 1 |
| Other cases | — | Not applicable: the frame has no roles, writes or failure state in the approved design |

**Data touched** — the riding types (names, ids); the bag count from the visitor's browser; the owner's words and contact details.

**Requirement SITE-002 — Home: the shop's leading bikes and why to trust it**

*As a* visitor, *I want* the home page to show me the riding, the bikes the shop leads with and the shop's promises, *so that* I can see what is for sale and trust it within one screen.

Behaviour:

1. The first screen: one `h1` in three lines of different weight — "Built for" (light), "Passes. Trails" (bold), "& rough ground" (light) — with a primary pill "See all bikes" and an arrow set on the last line, linking to the shop; a riding photograph fills the right half and runs off the screen edge (on a phone, edge to edge under the heading).
2. Under the heading, the proof block: up to three overlapping round photographs of live bikes that are in stock, the number of live bikes in stock followed by "bikes in stock", and the line "From a first hardtail to a carbon race bike — each one built and set up before it leaves." The block is not drawn when no live bike is in stock.
3. Then up to three tall tinted cards, one per live bike the owner marked "Lead with it" (the first three in the list's order), tinted pink, butter and lilac in that order. Each card shows a white pill with the riding type, the brand in light weight and the model in bold, the one-line fact, the bike's photograph in a window, and at its foot a translucent bar with the price and a round navy arrow button labelled "Choose a size <model>". The whole card opens the bike's page. The row is not drawn when no bike is marked.
4. Then the riding types as grey letter-spaced wordmarks between two hairlines, each linking to the shop narrowed to that type, and a last wordmark "Workshop" linking to the workshop prices on the About page.
5. Then the bike of the month: the live bike the owner marked "Bike of the month", on a lilac card beside a riding photograph, with the heading "Bike of / the month" set at the page's largest size, the bike's description, a "Travel" row and a "Weight" row with the bike's own values, and a primary button "See the bike" to its page. Not drawn when no bike is marked.
6. Then a grey band "Not sure / where to start?" with the line "Four kinds of mountain bike, and the riding each one is for.", one row per riding type (name in semibold, who it is for, where it rides best, a round arrow) linking to the shop narrowed to that type, and a riding photograph at the right.
7. Then a peach band of three promises, from the shop's own record: "Delivered across Vietnam" with the shop's in-town and elsewhere delivery terms; "<N> years of frame warranty" with "From the shop, on top of the maker's own warranty."; "A reply within <N> hours" with "Message us and a rider answers — about sizes, trails or an order."
8. Then "Favourite / picks": the line "More from the shop floor — from first-trail hardtails to carbon race bikes.", a row of chips — "All bikes" drawn selected and linking to the shop, then one chip per riding type linking to the shop narrowed to it — and up to six live bikes that are not marked "Lead with it", three to a row (one per row on a phone), each as a bike card (SITE-003).
9. Last, "What riders / say": the shop's average rating and the number of ratings ("<rating> from <count> ratings"), a link "All reviews" to the reviews on the About page, and the first three reviews as review cards (SITE-011).
10. A section whose records do not exist is not drawn; the page never shows an empty box.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Any window | Home opens | There is exactly one `h1`, with lines "Built for", "Passes. Trails" and "& rough ground" in light, bold and light weight |
| AC-2 | Home is open | The pill "See all bikes" on the heading's last line is pressed | The shop page opens |
| AC-3 | The records have 12 live bikes with stock above 0 | Home opens | The proof block shows three round bike photographs, "12" and "bikes in stock", and the line "From a first hardtail to a carbon race bike — each one built and set up before it leaves." |
| AC-4 | No live bike has stock above 0 | Home opens | The proof block is not drawn |
| AC-5 | A bike's stock is changed from 5 to 0 in the records | Home reloads | The in-stock count is one lower |
| AC-6 | Bikes talon-2, stance-0 and talon-e-plus-1 are the live bikes marked "Lead with it" | Home opens | Three cards are drawn in that order, tinted pink, butter and lilac, each with riding-type pill, brand and model, fact line, photograph, price and an arrow button labelled "Choose a size <model>" |
| AC-7 | The first tinted card | The card is pressed | The page of that bike opens |
| AC-8 | No bike is marked "Lead with it" | Home opens | No tinted-card row is drawn |
| AC-9 | Four riding types exist | The departments row is read | Four wordmarks link to /shop?category=<id> for their type and a fifth, "Workshop", links to /about#workshop |
| AC-10 | Bike trance-x-1 is marked "Bike of the month" | Home opens | The heading "Bike of / the month", that bike's description, its "Travel" and "Weight" values and a "See the bike" button to /shop/trance-x-1 are drawn |
| AC-11 | No bike is marked "Bike of the month" | Home opens | The bike-of-the-month section is not drawn |
| AC-12 | Four riding types exist | The "Not sure / where to start?" band is read | It shows the line "Four kinds of mountain bike, and the riding each one is for." and four rows, each with name, who it is for, where it rides best and an arrow to the shop narrowed to that type |
| AC-13 | The shop's record says 2 warranty years and 2 reply hours | The peach band is read | It reads "Delivered across Vietnam" with the delivery terms, "2 years of frame warranty" and "A reply within 2 hours" with their sub-lines |
| AC-14 | The shop's record is changed to 3 warranty years | Home reloads | The band reads "3 years of frame warranty" |
| AC-15 | Home has loaded | The "Favourite / picks" section is read | It shows the line, a selected "All bikes" chip, one chip per riding type, and at most six live bikes not marked "Lead with it" |
| AC-16 | The shop's record says rating 4.9 and 412 ratings | The reviews section is read | It reads "4.9 from 412 ratings", shows a link "All reviews" to /about#reviews, and three review cards |
| AC-17 | A 390 px wide window | Home opens | The heading, proof block and photograph stack in one column with the photograph edge to edge under the heading |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Missing records | A section's records do not exist (none marked, none in stock, no reviews) | That section is not drawn (AC-4, AC-8, AC-11) |
| Other cases | — | Not applicable: the page has no roles or writes, and the design draws no loading or error variant of it |

**Data touched** — bikes (status, "Lead with it", "Bike of the month", stock, name, brand, riding type, fact, photographs, price, description, travel, weight), riding types, reviews, the shop's record (delivery terms, warranty years, reply hours, rating, number of ratings), the owner's words.

**Requirement SITE-003 — The shop: every bike, narrowed by riding type, frame size, wheel, brand and price**

*As a* visitor, *I want* to see every bike on sale and narrow the list by riding type, frame size, wheel size, brand and price, *so that* I find the bikes that suit my riding and my budget.

Behaviour:

1. The head is a grey band running to both screen edges: a breadcrumb (a house icon labelled "Home", then "Bikes", then — when one riding type is chosen — that type's name), the `h1` "All the / bikes", the lead "Hardtails for first trails, full suspension for rough ground, carbon for racing and e-MTBs for the longest days.", the riding-type chips, and a landscape photograph off the right edge (a strip across the top on a phone).
2. The chips are "All types" and one per riding type. "All types" is drawn chosen when no type is chosen. Pressing a type chooses it (it is drawn filled and is `aria-pressed`) and narrows the list to it; pressing the chosen type again, or "All types", clears it. While exactly one type is chosen the `h1` becomes "<Type name> bikes" (name light, "bikes" bold) and the lead becomes that type's "who it is for" line.
3. Only bikes whose status is live are listed; a bike in draft is never listed.
4. Down the left of the list, a filter rail with four groups of tick boxes: "Frame size" (XS, S, M, L, XL), "Wheel size" ("29 inch", "Mixed 29 / 27.5 inch"), "Brand" (each brand among the live bikes, in alphabetical order) and "Price" ("Under 25 million ₫", "25–70 million ₫", "70–100 million ₫", "Over 100 million ₫"). A group with fewer than two options is not drawn. A group that has a choice shows the number chosen beside its title.
5. Within one group the choices are "any of"; across groups they all apply. A frame size matches a bike whose frame sizes include it. A price band includes its lower figure and excludes its upper figure: a bike priced exactly 25.000.000 ₫ is in "25–70 million ₫", not in "Under 25 million ₫".
6. Everything that narrows the list — riding types, the four groups, the search words and the sort order — lives in the address, so a filtered list is a link: reloading keeps it, opening the link elsewhere shows the same list, and coming back from a bike lands on the same list.
7. While anything narrows the list, an underlined "Clear filters" link under the groups clears every choice and the search words and keeps the sort order.
8. Above the grid: the number of matching bikes in large bold figures followed by "bikes", and a select labelled "Sort bikes" with "Recommended" (the records' own order, the default), "Price: low to high" and "Price: high to low" (by price as a number).
9. When search words are in the address (arriving from the search page's "Narrow by size, wheel and price"), a line "Results for “<words>”" with a "Clear search" button is shown, and the list is narrowed to bikes whose name, brand, riding type, one-line fact, description, frame, fork, drivetrain or travel contains the words, ignoring capitals and accents.
10. The grid shows 9 bikes, three to a row on 1024 px and up and one per row on a phone, each as a bike card. When more match, a button "Show more bikes" followed by the number still hidden, in brackets, shows 9 more; it disappears when all are shown.
11. While the bikes are loading, the grid is replaced by one line "Loading bikes…".
12. When no bike matches, the grid is replaced by a pale pink panel: "No bike matches all of that.", "Take one filter off — or tell us what you need and we will order it in." and a primary button "Clear filters" that clears every choice.
13. Under the list, a lilac band run off the right edge: "Between / two sizes?", "Send us your height and inside-leg and we will tell you which — or book a test ride." and a button "Ask about sizing" to the contact page.
14. **A bike card** shows: the bike's photograph on a white 4:3 tile with a hairline edge; a heart at its top right (SITE-006); the brand and riding type in small capitals ("GIANT · TRAIL"); the model name in semibold and under it the one-line fact; at the right the price in bold and, when the bike has an earlier price above its price, that earlier price struck through beneath it; the stock status (SITE-004); and a full-width outlined pill that reads "Choose a size" and opens the bike's page at its size choice, or reads "Ask when it is back" when the bike is out of stock. The card opens the bike's page; the model name underlines under the pointer. In a row the cards take turns in the photograph they lead with — the drivetrain close-up, the whole bike, the fork and cockpit — falling back to the whole bike where a bike has no such close-up, so two whole-bike photographs do not stand side by side.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | 14 live bikes | /shop opens | The `h1` reads "All the / bikes", the count reads "14 bikes", "All types" is drawn chosen, and the breadcrumb reads Home, Bikes |
| AC-2 | Four riding types | The chips are read | "All types" and four type chips are drawn |
| AC-3 | /shop is open | The chip "Trail" is pressed | It is drawn filled with `aria-pressed` true, only trail bikes are listed, the `h1` reads "Trail bikes", the breadcrumb gains "Trail" and the lead is the trail type's "who it is for" line |
| AC-4 | The "Trail" chip is chosen | The "Trail" chip is pressed again | The riding-type choice is cleared and all live bikes are listed |
| AC-5 | A bike has status draft | /shop opens | That bike is not listed and not counted |
| AC-6 | /shop is open | The rail is read | Four groups — "Frame size", "Wheel size", "Brand", "Price" — are drawn with the options named in Behaviour 4 |
| AC-7 | The live bikes are all from two brands | The "Brand" group is read | Both brands are listed alphabetically |
| AC-8 | All live bikes are from one brand | The rail is read | The "Brand" group is not drawn |
| AC-9 | /shop is open | The size "M" box is ticked | Only bikes whose frame sizes include M are listed, the count updates, "1" shows beside "Frame size" and "Clear filters" appears |
| AC-10 | The wheel boxes "29 inch" and "Mixed 29 / 27.5 inch" are ticked | The list is read | Bikes with either wheel size are listed |
| AC-11 | A bike costs exactly 25.000.000 ₫ | The price box "25–70 million ₫" is ticked | That bike is listed; with "Under 25 million ₫" ticked instead it is not |
| AC-12 | Riding type "Trail" and size "M" are chosen | The list is read | Only bikes that are trail bikes and have size M are listed |
| AC-13 | A list is narrowed and sorted | The page is reloaded | The same chips, boxes and sort order are chosen and the same bikes are listed |
| AC-14 | A bike is opened from a narrowed list | The browser goes back | The list is narrowed the same way |
| AC-15 | A narrowed list | "Clear filters" is pressed | All choices and the search words are cleared and the sort order is kept |
| AC-16 | /shop is open | "Price: low to high" is chosen in "Sort bikes" | Bikes are listed by price ascending as numbers; "Price: high to low" lists them descending; "Recommended" restores the records' order |
| AC-17 | /shop?q=stance is open | The head of the list is read | "Results for “stance”" and a "Clear search" button are shown and only bikes matching "stance" are listed |
| AC-18 | /shop?q=STANCE is open | The list is read | The same bikes as for "stance" are listed |
| AC-19 | More than 9 bikes match | /shop opens | 9 bikes are drawn and the button "Show more bikes (<remaining>)" shows the number not yet drawn |
| AC-20 | The button "Show more bikes (5)" is shown | It is pressed | Up to 9 more bikes are drawn and, when none are left hidden, the button disappears |
| AC-21 | The bikes are still loading | /shop is read | The single line "Loading bikes…" stands where the grid will be |
| AC-22 | The chosen filters match no bike | The page is read | The pale pink panel with "No bike matches all of that.", its sub-line and a primary "Clear filters" button is drawn instead of the grid |
| AC-23 | The panel from AC-22 | "Clear filters" is pressed | The filters are cleared and bikes are listed |
| AC-24 | /shop is open | The band under the list is read | It reads "Between / two sizes?", the sub-line, and a "Ask about sizing" button to /contact |
| AC-25 | A bike with stock 0 | Its card is read | The button reads "Ask when it is back" instead of "Choose a size" |
| AC-26 | A bike with an earlier price above its price | Its card is read | The price is bold at the right and the earlier price is struck through beneath it; a bike with no earlier price shows no second figure |
| AC-27 | A bike card | The "Choose a size" pill is pressed | The bike's page opens at its size choice |
| AC-28 | A row of three bike cards | The photographs are read | The first leads with the drivetrain close-up, the second with the whole bike, the third with the fork and cockpit |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| No match | Filters or search words match nothing | The pale pink panel of Behaviour 12 (AC-22) |
| Loading | The bikes have not arrived | The line "Loading bikes…" (AC-21) |
| Boundary | A price exactly on a band edge | Lower edge in, upper edge out (AC-11) |
| Other cases | — | Not applicable: the page has no roles or writes. The design draws no error variant of the list; the service contract's error envelope governs the API |

**Data touched** — bikes (status, riding type, brand, wheel, frame sizes, price, earlier price, stock, name, fact, description, frame, fork, drivetrain, travel, photographs), riding types.

**Requirement SITE-004 — Stock availability status, in the owner's words**

*As a* visitor, *I want* to see at a glance whether a bike can be had, *so that* I do not set my heart on a bike that is gone.

Behaviour:

1. Wherever a bike is shown with its stock — on its card, on its page, per size in its size guide — the status is a word and a mark together, never colour alone.
2. A bike with 4 or more in stock reads "In stock" with a done mark. A bike with 1 to 3 in stock reads "In stock" followed by "· <N> left". A bike with 0 (or no figure) reads "Out of stock" with an alert mark in the danger colour.
3. The words "In stock", "Out of stock" and "left" are the owner's, editable in the Studio.
4. The number comes from the bike's own record: when the owner changes a bike's stock, every page showing it shows the new status.
5. A bike with 0 in stock is not offered for the bag: its card button reads "Ask when it is back" and its page has no quantity stepper and no "Add to bag" (SITE-005).

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bike with stock 9 | Its card is read | It reads "In stock" with a done mark and no count |
| AC-2 | A bike with stock 2 | Its card or page is read | It reads "In stock · 2 left" |
| AC-3 | A bike with stock 0 | Its card or page is read | It reads "Out of stock" with an alert mark in the danger colour |
| AC-4 | A bike with stock 5 | The owner's record is changed to stock 0 and the card is reloaded | The card reads "Out of stock" |
| AC-5 | A bike's page | The size guide is read | Each size's Stock cell reads "In stock" or "Out of stock" |
| AC-6 | The owner has changed "Out of stock" to another phrase in the Studio | A bike with stock 0 is read | The owner's phrase is shown |
| AC-7 | A bike with stock 0 | Its card is read | The status mark is an icon beside the word, not a colour alone |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Boundary | Stock exactly 3 and exactly 4 | 3 reads "In stock · 3 left"; 4 reads "In stock" |
| Other cases | — | Not applicable: a single read with no roles, writes or failure state in the approved design |

**Data touched** — a bike's stock, the owner's words "In stock", "Out of stock", "left".

**Requirement SITE-005 — The bike page: photographs, price, stock, size, specification, ways to buy**

*As a* visitor, *I want* one bike's page with large photographs, its price, stock, specification and sizes, *so that* I can decide and either buy it or ask the shop.

Behaviour:

1. A breadcrumb: "Bikes" (to the shop), the riding type (to the shop narrowed to it), then the bike's name, which is not a link.
2. The photographs: the main picture at 4:3, large (on a phone edge to edge), and under it one thumbnail for each further photograph — the drivetrain close-up and the fork and cockpit close-up. The first picture is shown first; pressing a thumbnail shows it large (the chosen thumbnail is `aria-pressed`). A bike with one photograph shows no thumbnail strip. Under the pointer on a pointer device the main picture is looked into at twice its size.
3. A tinted head on the tint of the riding type's place in the list: a white pill with the riding type, the heart (SITE-006), the `h1` with the brand in light weight and the model in bold, the line "<fact> · <model year>", and — only when the bike has a rating — five stars (filled for the rating), the rating and "(<N> ratings)".
4. The price in large bold figures. When the bike has an earlier price above its price, the earlier price is struck through beside it and a pill "You save <difference>" is shown beside the stock status. Under the price, the stock status (SITE-004).
5. "Frame size" with a link "Size guide" to the size guide; one pill per frame size with the rider height it fits under it. A size the owner has marked out, and every size of a bike with 0 in stock, is drawn disabled and struck through. The size "M" is chosen at first if it is available, otherwise the first available size; pressing a pill chooses it (`aria-checked`).
6. When the bike is in stock: a quantity stepper from 1 up to the bike's stock (the minus disabled at 1, the plus disabled at the stock) and a primary button "Add to bag", which puts the chosen quantity of the chosen size in the bag. After it, the button reads "In your bag" with a check mark and a line "Added to your bag." with a link "See your bag" to the bag appears. Beside it, an outlined link "Ask about it" jumps to the ask band (SITE-013). When the bike is out of stock the stepper and "Add to bag" are not drawn and a solid "Ask when it is back" link to the ask band is drawn instead.
7. A bordered promise list: the shop's delivery terms, "<N> years of frame warranty from the shop", and "A reply within <N> hours, seven days a week".
8. "The / specification": a two-column list with, in this order, Frame, Fork, Rear shock, Motor and battery, Drivetrain, Brakes, Wheels, Tyres, Weight, Colour. A row whose value the bike lacks is not drawn.
9. "Which / size?": a table with columns "Size", "Rider height" and "Stock", one row per frame size, and under it the note "Heights are the maker's guide. Long legs or a short back can move you a size — send us both and we will check." Not drawn for a bike with no frame sizes.
10. Tabs: "What it is for" (the bike's description), "Owning it" (what it is like to own), "Delivery and payment" (rows "In town", "Elsewhere in Vietnam" with the courier charge, "Collect", "Payment", "Confirming", from the shop's record) and "Reviews (<n>)" (that bike's reviews as review cards, drawn only when it has some). A tab with nothing in it is left out; one tab alone is drawn as its content with no bar. The arrow keys move between tabs.
11. The ask band (SITE-013) and, under it, "More / <riding type>": up to three other live bikes of the same riding type as bike cards. Not drawn when there are none.
12. When no bike has the address's id: the line "That bike is no longer in the shop." and a primary button "See all bikes" to the shop.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | /shop/trance-x-1 | The page opens | The breadcrumb reads Bikes › Trail › the bike's name, the first two linking to /shop and /shop?category=trail |
| AC-2 | A bike with three photographs | The page opens | The main picture and two thumbnails are drawn; the first picture is shown large |
| AC-3 | A bike with three photographs | The second thumbnail is pressed | That photograph is shown large and that thumbnail is `aria-pressed` true |
| AC-4 | A bike with one photograph | The page opens | No thumbnail strip is drawn |
| AC-5 | A bike with a rating of 4.8 from 52 ratings | The head is read | A row of five stars filled to the rating, "4.8" and "(52 ratings)" are shown |
| AC-6 | A bike with no rating | The head is read | No stars row is drawn |
| AC-7 | A bike priced 18.990.000 ₫ | The page opens | "18.990.000 ₫" is shown in large bold figures |
| AC-8 | A bike priced 60.000.000 ₫ with an earlier price 66.000.000 ₫ | The page opens | The earlier price is struck through beside the price and a pill reads "You save 6.000.000 ₫" |
| AC-9 | A bike with no earlier price, or one equal to or below its price | The page opens | No struck-through price and no "You save" pill |
| AC-10 | A bike with sizes XS–XL, XL marked out | The page opens | Five pills show each size with its rider height; XL is disabled and struck through; M is chosen |
| AC-11 | A bike with sizes S and L only, both available | The page opens | S, the first available size, is chosen |
| AC-12 | A bike with stock 0 | The page opens | Every size pill is disabled |
| AC-13 | A bike in stock with M chosen and quantity 2 | "Add to bag" is pressed | The bag holds 2 of that bike in size M, the button reads "In your bag" with a check, and "Added to your bag." with a link "See your bag" to /bag appears |
| AC-14 | A bike with stock 3 | The quantity stepper is read | It goes from 1 to 3; the minus is disabled at 1 and the plus at 3 |
| AC-15 | A bike in stock | The page opens | An outlined "Ask about it" link beside "Add to bag" jumps to the ask band |
| AC-16 | A bike with stock 0 | The page opens | No stepper and no "Add to bag"; a solid "Ask when it is back" link to the ask band is drawn |
| AC-17 | The shop's record | The promise list is read | It shows the delivery terms, "2 years of frame warranty from the shop" and "A reply within 2 hours, seven days a week" for a record of 2 and 2 |
| AC-18 | A bike with all ten specification values | The specification is read | Ten rows in the order Frame, Fork, Rear shock, Motor and battery, Drivetrain, Brakes, Wheels, Tyres, Weight, Colour |
| AC-19 | A bike with no motor value | The specification is read | No "Motor and battery" row is drawn |
| AC-20 | A bike with sizes | The size guide is read | A table with "Size", "Rider height", "Stock" columns, one row per size, and the note "Heights are the maker's guide. Long legs or a short back can move you a size — send us both and we will check." |
| AC-21 | A bike with a description, an ownership text and one review | The tabs are read | "What it is for", "Owning it", "Delivery and payment" and "Reviews (1)" are drawn |
| AC-22 | A bike with no reviews | The tabs are read | No "Reviews" tab is drawn |
| AC-23 | The "Delivery and payment" tab | It is opened | Rows "In town", "Elsewhere in Vietnam" (with the courier charge), "Collect", "Payment" and "Confirming" show the shop's own text |
| AC-24 | A bike whose riding type has two other live bikes | The page is read to its foot | "More / <riding type>" shows those two as bike cards |
| AC-25 | A bike that is the only live bike of its type | The page is read to its foot | No "More" row is drawn |
| AC-26 | /shop/no-such-bike, after the records have answered | The page opens | "That bike is no longer in the shop." and a "See all bikes" button to /shop are shown |
| AC-27 | A bike's price is changed in the records | The page is reloaded | The new price is shown |
| AC-28 | A bike with the size S marked out of stock | The size guide is read | The S row's Stock cell reads "Out of stock" |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not found | The address's id matches no bike | The line and button of Behaviour 12 (AC-26); see also SITE-016 for the moment before the answer |
| Out of stock | Stock is 0 | No way to add to the bag; the ask path is offered (AC-16) |
| All sizes marked out but stock above 0 | The owner marked every size out and left stock above 0 | The design draws every pill disabled and still draws "Add to bag" with no size; the order is then refused at checkout by the order rules (ORD-004). Not specified further |
| Draft bike | A bike in draft is opened by its address | Not decided by the brief or the design: see §7, open question 2 |

**Data touched** — a bike's every field the page reads (name, brand, riding type, year, price, earlier price, fact, stock, frame sizes, sizes out, rider heights, the ten specification values, rating, number of ratings, description, ownership text, three photographs), the bike's reviews, the shop's record, riding types.

**Requirement SITE-006 — Keep a bike with the heart**

*As a* visitor, *I want* to keep a bike for later with a heart, *so that* I can find my favourites again on this browser.

Behaviour:

1. A 32 px round white heart sits at the top right of every bike card's photograph and in the tinted head of a bike's page. It is drawn in the soft ink colour, in ink under the pointer.
2. Pressing it keeps the bike: the heart is filled in the accent colour and is `aria-pressed` true. Pressing it again removes it. Pressing the heart does not open the bike's page.
3. The choice is remembered in this browser across reloads. Where the browser refuses storage the heart lasts for the visit.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bike card | The heart is pressed | The heart is filled and `aria-pressed` is true, and the bike's page has not opened |
| AC-2 | A kept bike | Its heart is pressed | The heart is empty again |
| AC-3 | A kept bike | The page is reloaded | The bike's heart is still filled on its card and on its page |
| AC-4 | A bike's page | The head is read | The same heart is drawn in the tinted head |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Storage refused | The browser refuses local storage | The heart stays filled for this visit and is lost on reload |

**Data touched** — the ids of the bikes kept, in the visitor's own browser only.

**Requirement SITE-007 — The bag and the bag drawer**

*As a* visitor, *I want* a bag that holds the bikes I chose and shows what they come to, *so that* I can check them before I check out.

Behaviour:

1. The bag is kept in this browser and survives a reload. Each line keeps the bike's name, price, photograph and chosen size as they were when it was added. Adding the same bike in the same size again adds to that line's quantity; the same bike in another size is another line.
2. The bag page's head is a grey band: the `h1` "Your / bag", the lead "Check the bike and the size. How it reaches you is chosen on the next step.", and a photograph of a bike's bars off the right edge (a strip on a phone).
3. Each line shows the bike's photograph, its name (a link to the bike's page), "Size <size> · <price>", a quantity stepper, the line total in bold, and a round remove button labelled "Remove <name>". The stepper's minus is disabled at 1; the remove button takes the line out.
4. Beside the lines (below them on a phone) a pale pink summary card: "Bikes" with the subtotal, "Delivery" with "Chosen at checkout", a translucent total row "Total" with the total in large bold figures, a primary button "Check out" to the checkout page, and a list of the shop's delivery terms, "<N> years of frame warranty from the shop" and "We call to confirm before anything is charged."
5. An empty bag shows a pale pink panel: "Your bag is empty.", "Start from the riding you do — each kind of bike is one click away.", one chip per riding type to the shop narrowed to it, and a primary button "See all bikes".
6. The bag drawer, opened from the header's bag button on any page, is a panel from the right edge over a dimmed page: the title "Your bag", a close button, the lines (photograph, name, size, line total, quantity stepper, remove), and at its foot the total and a primary "Check out" link to the checkout page. Empty, it shows only "Your bag is empty." and no footer. Escape closes it, the focus stays inside it while open, and the page behind does not scroll.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bike in size M was added once | The same bike in size M is added again | One line shows quantity 2 |
| AC-2 | A bike in size M is in the bag | The same bike in size L is added | Two lines exist |
| AC-3 | A bag with items | The page is reloaded | The same lines are in the bag |
| AC-4 | A bag with one line of 2 × 18.990.000 ₫ | /bag opens | The line shows "Size M · 18.990.000 ₫", the line total 37.980.000 ₫, and the summary's "Bikes" and "Total" both read 37.980.000 ₫ |
| AC-5 | A line with quantity 1 | The stepper is read | The minus is disabled |
| AC-6 | A line | "+" is pressed | Quantity and line total and the totals rise |
| AC-7 | A line named "Giant Talon 2" | Its remove button is pressed | The line is gone; the button's label was "Remove Giant Talon 2" |
| AC-8 | /bag is open with items | The summary card is read | It shows "Bikes", "Delivery — Chosen at checkout", "Total", a "Check out" button to /checkout, the delivery terms, the warranty line and "We call to confirm before anything is charged." |
| AC-9 | /bag is open | The head is read | The `h1` reads "Your / bag" and the lead reads "Check the bike and the size. How it reaches you is chosen on the next step." |
| AC-10 | The bag is empty | /bag opens | The panel "Your bag is empty.", its sub-line, four riding-type chips and the "See all bikes" button are shown, with no summary card |
| AC-11 | Any page, the bag holds items | The header's bag button is pressed | The drawer opens over a dimmed page with the title "Your bag", the lines, the total and "Check out" |
| AC-12 | The drawer is open | Escape is pressed | It closes |
| AC-13 | The bag is empty | The header's bag button is pressed | The drawer shows "Your bag is empty." and no total or "Check out" |
| AC-14 | A bike's price changes in the records after it was added | /bag opens | The line still shows the price it had when added |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Storage refused | The browser refuses local storage | The bag lasts for the visit |
| Price changed since added | The shop changed the price after the bag line was made | The bag shows the old figure; the order is priced from the shop's current price by the server (ORD-002) and the confirmation shows that total |

**Data touched** — in the visitor's browser: a line per bike and size (bike id, name, price, photograph, size, quantity).

**Requirement SITE-008 — The checkout page: who, how it reaches you, how you pay**

*As a* visitor with bikes in my bag, *I want* a checkout in three steps, *so that* I can tell the shop who I am, where the bike goes and how I will pay.

Behaviour:

1. The head is a grey band: the `h1` "Check / out", the lead "Three steps. We confirm by phone before anything is charged.", and a close-up photograph off the right edge.
2. Left, one form of three numbered groups. **1 "Who it is for":** fields "Your name" and "Phone number", both required. **2 "How it reaches you":** three option cards — "Delivered in town" (with the shop's in-town terms), "Courier, anywhere in Vietnam" (with the shop's courier terms and the courier charge in bold when it is above 0) and "Collect from the shop" (with the shop's collection terms); "Delivered in town" is chosen at first. Unless the customer collects, a required field "Delivery address" shows under the cards; for "Collect from the shop" it is not drawn. **3 "How you pay":** three option cards — "Bank transfer" ("We send a VietQR code when we call to confirm."), "Cash on delivery" ("Pay when the bike arrives. In town only.") and "At the shop" ("Card or cash when you collect."); "Bank transfer" is chosen at first; "Cash on delivery" is disabled unless the delivery is "Delivered in town", and choosing another delivery while it is chosen switches the payment to "Bank transfer". Then an optional field "A note for the shop (optional)". An option card is chosen by pressing it and shows a filled ring with a check and a pink tint.
3. A hidden field named "website" is part of the form and is not visible or reachable by a person.
4. A butter-coloured panel closes the form: the shop's confirming sentence with a phone icon, the failure and "not connected" lines when they apply (SITE-008 failure table), and the primary button "Place order · <total>". The button is disabled while the order is being sent and while the bag is empty.
5. Right, the summary "Your order": a line per bike (photograph, name, "Size <size> · <quantity> × <price>", line total), then "Bikes" with the subtotal, "Delivery" with the courier charge or "Free", the total in large bold figures, and "Every bike is built, checked and set up for you before it leaves the shop." With an empty bag it reads "Your bag is empty." with a link "See all bikes".
6. The courier charge is the shop's own figure and is added to the total only while "Courier, anywhere in Vietnam" is chosen. The total on the button and in the summary is the subtotal plus that charge.
7. What pressing "Place order" does — sending the order, the refusal sentences, the confirmation — is specified in the `orders` module (ORD-001 to ORD-008).

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bag with items | /checkout opens | The `h1` reads "Check / out", the lead reads "Three steps. We confirm by phone before anything is charged.", and groups 1, 2 and 3 are numbered "Who it is for", "How it reaches you", "How you pay" |
| AC-2 | /checkout is open | Group 1 is read | "Your name" and "Phone number" are required fields |
| AC-3 | /checkout opens | Group 2 is read | Three cards are drawn, "Delivered in town" chosen, each with the shop's text from its record |
| AC-4 | The shop's courier charge is 350.000 ₫ | Group 2 is read | The courier card shows "350.000 ₫" in bold |
| AC-5 | "Delivered in town" is chosen | The page is read | A required "Delivery address" field is drawn |
| AC-6 | "Collect from the shop" is chosen | The page is read | No "Delivery address" field is drawn |
| AC-7 | /checkout opens | Group 3 is read | Three cards "Bank transfer", "Cash on delivery", "At the shop" with their lines are drawn, "Bank transfer" chosen, and the optional field "A note for the shop (optional)" below |
| AC-8 | "Courier, anywhere in Vietnam" is chosen | The payment cards are read | "Cash on delivery" is disabled |
| AC-9 | "Cash on delivery" is chosen | "Collect from the shop" is chosen | The payment becomes "Bank transfer" |
| AC-10 | A bag worth 18.990.000 ₫ and "Delivered in town" chosen | The page is read | The button reads "Place order · 18.990.000 ₫" and the summary's delivery reads "Free" |
| AC-11 | The same bag, courier charge 350.000 ₫ | "Courier, anywhere in Vietnam" is chosen | The button and the summary's total read 19.340.000 ₫ and the delivery row reads "350.000 ₫" |
| AC-12 | A bag with items | The summary is read | Each line shows photograph, name, "Size M · 1 × 18.990.000 ₫" and its line total, and the note "Every bike is built, checked and set up for you before it leaves the shop." is under the total |
| AC-13 | The bag is empty | /checkout opens | The summary reads "Your bag is empty." with a "See all bikes" link and "Place order" is disabled |
| AC-14 | The form is being sent | The button is read | It is disabled and reads "Sending…" |
| AC-15 | /checkout is open | The page is read | The shop's confirming sentence with a phone icon is above the button |
| AC-16 | /checkout is open | The form is inspected | A field "website" exists that a person cannot see or tab to |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Required field empty | The name, phone or (when shown) address is empty | The browser's own required-field check stops the submit; nothing is sent |
| Failed send | The order is refused or cannot be sent | A line with the sentence in the danger colour under the fields, in the butter panel (`role="alert"`), the button enabled again (ORD-007) |
| Not connected | The site has no line to the owner | The line "This form is not connected yet — please try again later." in the soft ink colour above the button |
| Empty bag | The bag has no lines | AC-13 |

**Data touched** — the bag's lines; the shop's record (in-town, courier and collection terms, courier charge, confirming sentence); what the visitor types (name, phone, address, note) and chooses (delivery, payment).

**Requirement SITE-009 — The order confirmation page**

*As a* visitor who has placed an order, *I want* to see my order number, what I ordered and what happens next, *so that* I know the shop has it and what to expect.

Behaviour:

1. With an order saved in this browser: a pill "Order received"; the `h1` "Thank you. / Your bike / is on hold"; the sentence "We will call <name> within <N> hours of opening to confirm the size and delivery." (N is the shop's reply hours); "Order number" with the number in large bold figures and the date and time the order was placed; a riding photograph off the right edge.
2. A navy band of four steps in a row, the first lit in butter: "Order received — Your bike is held for you.", "We call you — Size, delivery and payment confirmed.", "Built and checked — Assembled, adjusted and test-ridden.", "On its way — Delivered, sent or ready to collect."
3. "What you ordered": each line with photograph, name, "Size <size> · <quantity> × <price>" and line total, then "Total" with the order's total in large bold figures.
4. A bordered box: "How it reaches you" with the delivery method's name and, when there is an address, " — <address>"; "How you pay" with the payment method's name; then "Questions about this order?" with a primary button "Message us" to the contact page.
5. With no order saved in this browser: a navy band to both screen edges with the `h1` "No order / yet", the line "Orders you place on this browser show here." and a white button "See all bikes", beside the same photograph.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | An order saved in this browser | /checkout/done opens | The pill "Order received", the `h1` "Thank you. / Your bike / is on hold" and the photograph are shown |
| AC-2 | An order for "Minh" and the shop's reply hours 2 | The page is read | "We will call Minh within 2 hours of opening to confirm the size and delivery." is shown |
| AC-3 | An order with number DEO-261116-4821 placed 2026-11-16 at 09:40 | The page is read | "Order number", "DEO-261116-4821" and the placed date and time are shown |
| AC-4 | The page with an order | The band is read | Four steps with the titles and lines of Behaviour 2, the first lit in butter |
| AC-5 | An order of 1 × a 66.490.000 ₫ bike in size M | "What you ordered" is read | The line shows its photograph, name, "Size M · 1 × 66.490.000 ₫" and 66.490.000 ₫, and "Total" reads 66.490.000 ₫ |
| AC-6 | An order with delivery "Delivered in town" and address "12 Hùng Vương" | The box is read | "How it reaches you" reads "Delivered in town — 12 Hùng Vương" |
| AC-7 | An order with delivery "Collect from the shop" and no address | The box is read | "How it reaches you" reads "Collect from the shop" with no dash |
| AC-8 | An order paid by "Bank transfer" | The box is read | "How you pay" reads "Bank transfer" |
| AC-9 | The page with an order | "Message us" is pressed | The contact page opens |
| AC-10 | No order saved in this browser | /checkout/done opens | The navy band with "No order / yet", "Orders you place on this browser show here." and a white "See all bikes" button to /shop is shown |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| No order | None saved in this browser | AC-10 |
| Other cases | — | Not applicable: the page only reads an order already saved; the order itself and its refusals are in the `orders` module |

**Data touched** — the last order saved in this browser (number, placed time, name, delivery, address, payment, lines, total); the shop's reply hours.

**Requirement SITE-010 — The gallery, shown large**

*As a* visitor, *I want* to see the riding these bikes are bought for in large photographs, *so that* I can picture where I would ride.

Behaviour:

1. The `h1` "Where / these bikes / go" with the lead "Climbs to the pass, rock on the way down, roots under the trees — the riding our customers buy their bikes for.", the number of photographs followed by "photographs", and one outlined chip per riding type linking to the shop narrowed to it.
2. The first photograph is drawn wide beside the heading and runs off the right edge; the rest follow two to a row at 768 px and up (one per row on a phone), every frame the same landscape shape; a last odd photograph runs the full width. Photographs are large: each is at least half the page's width at 768 px and up, and full width on a phone.
3. Every photograph has a caption under a hairline: its title in semibold, its caption line, and — when it has a riding type — an outlined link with that type's name to the shop narrowed to it.
4. A closing band: "Seen / your kind / of riding?" with a primary button "See the bikes for it" to the shop.
5. An address ending in `#shot-<id>` scrolls to that photograph.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | 7 photographs and 4 riding types | /gallery opens | The `h1` "Where / these bikes / go", the lead, "7 photographs" and four type chips are shown |
| AC-2 | /gallery is open | The grid is read | The first photograph is drawn wide at the right of the heading and six more follow, two to a row, the last of an odd count at full width |
| AC-3 | A 390 px wide window | /gallery opens | Every photograph is drawn at the full width of the screen, one per row |
| AC-4 | A photograph "The long climb to the pass" of type Hardtail | Its caption is read | It shows the title, the caption line and a "Hardtail" link to /shop?category=hardtail |
| AC-5 | A photograph with no riding type | Its caption is read | No type link is drawn |
| AC-6 | /gallery is open | The foot is read | "Seen / your kind / of riding?" and a "See the bikes for it" button to /shop |
| AC-7 | The address /gallery#shot-ridge-view | The page opens | The page scrolls to that photograph |
| AC-8 | A photograph's title is edited in the records | /gallery reloads | The new title is shown |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not applicable | — | Not applicable: a single read with no roles, writes or failure state in the approved design |

**Data touched** — gallery photographs (image, title, caption, riding type), riding types.

**Requirement SITE-011 — About: the story, the team, the prices, what riders say, the questions**

*As a* visitor, *I want* to know who is behind the shop, what the workshop costs, what riders say and the answers to what I would ask before buying, *so that* I can trust the shop.

Behaviour:

1. The head: the `h1` "Bikes we / ride. Advice / we would take", the lead "We are a small shop of riders. Every bike is built in our workshop and set up for the person riding it.", the story "We sell the bikes we would choose ourselves, tell you when a cheaper one will do, and look after it when it comes back muddy.", three figures — "Open since" (the year), "Bikes built" (the number, counting up) and "Average rating" — from the shop's record, and the workshop photograph, large, off the right edge.
2. "The / people" with the line "Four riders who build, fit and deliver every bike." and one tinted card per team member: a white pill with the role, the name in large bold type, a line about them, and — when given — "Rides: <what they ride>".
3. "Workshop / prices" (anchor `workshop`) with the line "Every bike we sell is built and set up free. After that, this is what work costs." and a pill "First service free within <N> months"; then a table with columns "Job", "Time" and "Price", one row per workshop job with its name, what is done, how long and its price. A job priced 0 reads "Included". The "Time" column is not drawn below 640 px.
4. "What riders / say" (anchor `reviews`) with "<rating> from <count> ratings"; the first six reviews as review cards; and — when there are more — a button "Show all reviews" with the total in brackets that shows all of them. **A review card** shows five stars filled to the rating (`role="img"`, label "<n> out of 5"), a quote mark, the review's text, then under a hairline the first name in semibold, the bike's name when the review is about a bike, and the date as "16 Oct 2026".
5. "Before / you buy" (anchor `faq`) with a primary button "Ask us anything" to the contact page, and one row per question as a disclosure: the question with a plus mark that turns when open, and its answer when open. Each question's anchor is `faq-<id>`; each job's is `job-<id>`.
6. An address ending in an anchor above scrolls to it.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The shop's record: founded 2016, 3200 bikes, rating 4.9 | /about opens | The `h1`, lead, story, "Open since 2016", "Bikes built 3200" and "Average rating 4.9" and the workshop photograph are shown |
| AC-2 | Four team members | The people section is read | "The / people", the line and four tinted cards are shown, each with role pill, name, line and "Rides: …" |
| AC-3 | A team member with no "rides" | Their card is read | No "Rides:" line is drawn |
| AC-4 | Six workshop jobs | The workshop section is read | The heading, the line, the pill "First service free within 3 months" and a table of six rows are shown |
| AC-5 | A job priced 250000 and one priced 0 | The table is read | They read "250.000 ₫" and "Included" |
| AC-6 | A 390 px wide window | The table is read | The "Time" column is not drawn |
| AC-7 | 12 reviews | The reviews section is read | The heading, "4.9 from 412 ratings" for the shop's record 4.9 and 412, six review cards and a button "Show all reviews (12)" are shown |
| AC-8 | The button "Show all reviews (12)" | It is pressed | All 12 cards are shown and the button is gone |
| AC-9 | A review of 4 stars by "Thảo" about a bike on 2026-09-14 | Its card is read | Five stars with four filled and label "4 out of 5", the text, "Thảo", the bike's name and "14 Sep 2026" |
| AC-10 | A review with no bike | Its card is read | Only the name and the date follow the text |
| AC-11 | Eight questions | The questions section is read | "Before / you buy", a primary "Ask us anything" button to /contact and eight collapsed rows |
| AC-12 | A collapsed question | It is pressed | Its answer shows and the plus mark turns |
| AC-13 | The address /about#faq-pay | The page opens | The page scrolls to that question |
| AC-14 | The address /about#workshop | The page opens | The page scrolls to the workshop prices |
| AC-15 | A service's price is changed in the records | /about reloads | The new price is shown |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Missing records | A section's records do not exist | That section is not drawn |
| Other cases | — | Not applicable: no roles, writes or failure variant is drawn |

**Data touched** — the shop's record (year, bikes built, rating, number of ratings, free service months), team members, workshop jobs, reviews, questions.

**Requirement SITE-012 — Contact page and the message form**

*As a* visitor, *I want* to send the shop a message with my name, my phone or email and what I want to say, *so that* the shop can answer me.

*As the* shop owner, *I want* what is sent to reach me in Agent Team, *so that* I read it and reply.

Behaviour:

1. The `h1` "Ask us / anything.", the lead "About a bike, a size, a test ride or an order. Tell us your height and where you ride, and we answer with the bike that fits.", a butter pill "We reply within <N> hours" (the shop's reply hours), then the form. On a phone a photograph leads, edge to edge, above the heading.
2. The form's fields, in order: "Your name" (required), "Phone or email" (required), "Which bike (if you know)", "Your height in cm" (numbers), "Your message" (four lines). A hidden field named "website" is not visible or reachable by a person. The primary button reads "Send message".
3. Pressing it sends the visitor's name, phone or email, message and the bike and height (when typed), with the page they were on, to the owner's inbox in Agent Team. The owner reads each as one message showing those values.
4. While sending the button reads "Sending…" and is disabled. When sent, the form is replaced by one line "Thank you — we have your message and will answer soon." When the sending fails, a line with the sentence the sender returned shows under the fields (`role="alert"`) and the form keeps what was typed. When the site has no line to the owner, the line "This form is not connected yet — please try again later." shows above the button.
5. A box beside the photograph lists, each with its icon, the shop's phone, email, address and opening hours — only those the shop has given; none is drawn when none is given. The phone and email are links. When an address is given a map of it is drawn under everything.
6. Three tinted cards: "Which size?" ("Send your height and inside-leg and we tell you which frame fits."), "A test ride" ("Name the bike and a day, and we have it ready on our loop.") and "Workshop work" ("Book a service, a fit or a suspension job.").

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The shop's reply hours are 2 | /contact opens | The `h1` "Ask us / anything.", the lead and the butter pill "We reply within 2 hours" are shown |
| AC-2 | /contact is open | The form is read | The fields "Your name", "Phone or email", "Which bike (if you know)", "Your height in cm", "Your message" and the button "Send message" are shown in that order; the first two are required |
| AC-3 | The name or "Phone or email" is empty | "Send message" is pressed | The browser stops the submit and nothing is sent |
| AC-4 | Name "Lan", phone "0901234567" and message "Which size for 165 cm?" | "Send message" is pressed | The owner's Agent Team inbox receives one message with that name, that phone and that message, and the page /contact |
| AC-5 | A bike "Stance 0" and height "165" are typed | "Send message" is pressed | The message in Agent Team also shows the bike and the height |
| AC-6 | The message is being sent | The button is read | It is disabled and reads "Sending…" |
| AC-7 | The message was accepted | The page is read | The form is replaced by "Thank you — we have your message and will answer soon." |
| AC-8 | The sender refuses the message with a sentence | The page is read | The sentence is shown under the fields with `role="alert"` and the typed values are still in the fields |
| AC-9 | The site has no line to the owner | /contact opens | "This form is not connected yet — please try again later." is shown above the button |
| AC-10 | The hidden field "website" is filled | The message is sent | The request carries the filled value in its `website` field, which is how Agent Team recognises and discards a bot's message |
| AC-11 | The shop has given no phone, email, address or hours (as shipped) | /contact opens | No details box and no map are drawn |
| AC-12 | The shop has entered a phone and an email | /contact opens | The box lists them with their icons as `tel:` and `mailto:` links |
| AC-13 | The shop has entered an address | /contact opens | The address is listed and a map is drawn under the page |
| AC-14 | /contact is open | The row of cards is read | "Which size?", "A test ride" and "Workshop work" with their lines are shown |
| AC-15 | A 390 px wide window | /contact opens | The photograph is edge to edge above the heading |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Name or phone/email empty | The browser's required-field check stops the submit (AC-3) |
| Upstream failure | The line to Agent Team fails | The sender's sentence in a `role="alert"` line, the typed values kept (AC-8) |
| Not connected | The site has no line to the owner | The line of Behaviour 4 (AC-9) |
| Spam | The hidden field is filled | Not shown as a person's message (AC-10) |

**Data touched** — the visitor's name, phone or email, message, bike, height and the page; the shop's reply hours and contact details.

**Requirement SITE-013 — Ask the shop to call, on a bike's page**

*As a* visitor, *I want* to leave my number and my height on a bike's page, *so that* the shop calls me back with the right size and closes the sale.

*As the* shop owner, *I want* the bike, the size and the height attached to what is sent, *so that* I can call back with an answer, not a question.

Behaviour:

1. The ask band is a grey band run off the right edge with the heading "Ask about the / <model>" (model in bold) and the line "Leave your number and your height. We call back with the right size and when you can ride it." — for a bike with 0 in stock, "This one has sold out. Leave your number and we call you when the next one arrives." — and the form.
2. The form's fields: "Your name" (required), "Phone or email" (required), "Your height (cm)" (a number from 120 to 220), a select "Size you are thinking of" listing the bike's frame sizes with the placeholder "Not sure yet", and "Anything else — where you ride, a test ride". A hidden field "website" as in SITE-012. The button reads "Ask the shop to call".
3. Pressing it sends name, phone or email, the message (or, when empty, the bike's brand and model), and the bike, the size ("Not sure yet" when none was chosen) and the height, to the owner's inbox in Agent Team.
4. While sending the button reads "Sending…" and is disabled. When sent, the form is replaced by "Thank you — we have your details and will call you back with an answer." On failure a line with the sender's sentence shows under the fields in the danger colour (`role="alert"`) and the values are kept. With no line to the owner, "This form is not connected yet — please try again later." shows.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bike in stock | Its page is read | The ask band reads "Ask about the / <model>" and "Leave your number and your height. We call back with the right size and when you can ride it." |
| AC-2 | A bike with stock 0 | Its page is read | The band's line reads "This one has sold out. Leave your number and we call you when the next one arrives." |
| AC-3 | A bike with sizes S, M, L | The size select is opened | It lists S, M and L and its placeholder reads "Not sure yet" |
| AC-4 | The form is read | The fields are listed | "Your name", "Phone or email", "Your height (cm)", "Size you are thinking of", "Anything else — where you ride, a test ride" and the button "Ask the shop to call" |
| AC-5 | A height of 250 is typed | The form is submitted | The browser's range check stops it; nothing is sent |
| AC-6 | Name "Lan", phone "0901234567", size M, height 165 and no message, for bike "Giant Stance 0" | "Ask the shop to call" is pressed | Agent Team receives a message from Lan with that phone whose text is the bike's name and whose details show the bike, size M and height 165 |
| AC-7 | No size is chosen | The form is submitted | The details show the size as "Not sure yet" |
| AC-8 | The form is being sent | The button is read | It is disabled and reads "Sending…" |
| AC-9 | The details were accepted | The page is read | The form is replaced by "Thank you — we have your details and will call you back with an answer." |
| AC-10 | The sender refuses with a sentence | The page is read | The sentence is shown under the fields in the danger colour and the values stay |
| AC-11 | The site has no line to the owner | The page is read | "This form is not connected yet — please try again later." is shown |
| AC-12 | A bike in stock | "Ask about it" in the buy area is pressed | The page jumps to this band |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Name or phone/email empty, or a height outside 120–220 | The browser's own check stops the submit; nothing is sent (AC-5) |
| Upstream failure | The line to Agent Team fails | AC-10 |
| Not connected | No line to the owner | AC-11 |

**Data touched** — name, phone or email, height, size, message, the bike's brand and model; the sizes of the bike.

**Requirement SITE-014 — Search across the site**

*As a* visitor, *I want* to search the bikes, riding types, workshop jobs, buying questions and gallery from one field, *so that* I find what I need without knowing where it lives.

Behaviour:

1. The head: the `h1` "Search / the shop", the lead "Bikes and their specifications, workshop jobs and prices, the questions riders ask before buying, and the gallery.", a pill-shaped field with the hint "A model, a job or a question — Stance, brake bleed, size" and a dark "Search" button set inside it, and a riding photograph off the right edge. What is typed and the chosen kind of result live in the address, so a search is a link.
2. With nothing typed: a pink panel "What are you looking for?" with the line "A bike, a part of one, a workshop job or a question about buying — try one of these." and six suggestion chips — "Stance", "E-MTB", "Brake bleed", "Size", "Delivery", "Carbon" — each searching its word; and a grey panel "Or start from / a riding type" with a row per riding type (name, where it rides best, a round arrow) to the shop narrowed to it.
3. With words typed, matching ignores capitals and accents: "cua lo" finds "Cửa Lò". Results come in five kinds: Bikes (only live bikes; name, brand, riding type, one-line fact, description, frame, fork, drivetrain, travel), Riding types (name, who it is for, where it rides best), Workshop (job name, what is done), Questions (question, answer), Gallery (title, caption, riding type).
4. Beside the heading the total number of results followed by "results". A row of chips "Show results from": "Everything" with the total, then one chip for each kind that has results — "Bikes", "Riding types", "Workshop", "Questions", "Gallery" — with its count; the chosen chip is filled. Choosing a kind shows only that kind. A kind asked for in the address that has no results for these words shows everything.
5. A kind with results is a block headed by its name in bold and its count in light weight, and a link to where it lives at the right: "Narrow by size, wheel and price" (to the shop with the same words), "All bikes by type", "All workshop prices", "All questions", "The whole gallery". Bikes: up to six bike cards. Riding types: rows as in the grey panel. Workshop: a table "Job", "Time", "Price" with a job priced 0 as "Included". Questions: the question (a link to it on the About page) beside its answer. Gallery: photographs with title and caption linking to them on the gallery page.
6. When nothing is found: a pink panel "Nothing in the shop matches “<words>”" with "Try a shorter word or another spelling — or send us the question and a rider answers.", a primary button "Ask the shop" to the contact page and an outlined "See all bikes" to the shop.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | /search opens with nothing typed | The page is read | The `h1` "Search / the shop", the field with its hint and "Search" button, the pink panel with six suggestions and the grey panel with four riding types are shown, and no result count or kind chips |
| AC-2 | The suggestion "Brake bleed" | It is pressed | The page shows results for "brake bleed" |
| AC-3 | /search?q=bleed | The page is read | The "Workshop" block lists "Brake bleed" and the count beside the heading counts every result |
| AC-4 | A search that has results in Bikes and Questions | The chips are read | "Everything <total>", "Bikes <n>" and "Questions <n>" are drawn and no chip for a kind with 0 |
| AC-5 | "Workshop" is chosen | The page is read | Only the Workshop block is shown; the chip is filled; the address holds the chosen kind |
| AC-6 | The address asks for kind "gallery" and the words have no gallery result | The page opens | Every kind with results is shown and "Everything" is chosen |
| AC-7 | The bikes' records hold a live bike "Stance 0" | /search?q=STANCE and /search?q=stance | The same bikes are found; a draft bike is never found |
| AC-8 | A record whose text holds "Cửa Lò" | /search?q=cua lo | That record is found |
| AC-9 | /search?q=size | The page is read | The question "How do I know which frame size to choose?" is found under "Questions"; and where the Bikes block is drawn it lists at most six bike cards and carries the link "Narrow by size, wheel and price" to /shop?q=size |
| AC-10 | The Workshop block | The table is read | Its columns are "Job", "Time", "Price", a job priced 0 reads "Included" |
| AC-11 | A search with no result in any kind | The page is read | The pink panel "Nothing in the shop matches “<words>”", its line, a primary "Ask the shop" to /contact and an outlined "See all bikes" to /shop |
| AC-12 | Words typed in the field | Enter is pressed | The address holds the words and the results show |
| AC-13 | The Gallery block | A photograph is read | It shows the title, the caption and links to /gallery#shot-<id> |
| AC-14 | The Questions block | A question is read | It links to /about#faq-<id> |
| AC-15 | The Riding types block | Its link is read | "All bikes by type" to /shop |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Nothing found | No kind has a result | The panel of Behaviour 6 (AC-11) |
| Nothing typed | The field is empty | The suggestions and riding types (AC-1) |
| Other cases | — | Not applicable: no roles or writes; the design draws no error variant of the results |

**Data touched** — bikes, riding types, workshop jobs, questions, gallery photographs.

**Requirement SITE-015 — Made for the phone first**

*As a* visitor on a phone, *I want* every page to fit my screen, *so that* I can browse and order with my thumb.

Behaviour:

1. The design has two layouts and the switch is 1024 px. Below it: the header nav and sign-in icon give way to the menu toggle; the contact buttons are drawn (when numbers are given); the bike page stacks its photographs above its buy area; filter groups on the shop fold behind their titles and open by tapping them; the product and bag pictures run edge to edge. From it: the nav, search field and sign-in icon, the filter rail beside the list, and the product page with the photographs left and the buy area right.
2. At 768 px, cards and sections use up to three columns where the design draws them; below 640 px bikes, reviews and promises are one per row, and small tiles two per row.
3. No page scrolls sideways at any width from 360 px up.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Each storefront page at 360 px, 390 px and 768 px wide | The page is scrolled | There is no horizontal scroll bar and nothing is cut off at the right edge |
| AC-2 | A 390 px wide window | /shop opens | The filter groups are drawn folded behind their titles; tapping "Frame size" opens it |
| AC-3 | A 1280 px wide window | /shop opens | The filter groups are drawn open down the left of the list and their titles are not toggles |
| AC-4 | A 390 px wide window | A bike's page opens | The photographs are above the head, price and buy area, in one column |
| AC-5 | A 1280 px wide window | A bike's page opens | The photographs are at the left and the head, price and buy area at the right |
| AC-6 | A 390 px wide window | /shop opens | Bike cards are one per row |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not applicable | — | Not applicable: a layout requirement with no roles, writes or failure state |

**Data touched** — none.

### 4.2 Bike page waits for its record

**Requirement SITE-016 — The bike page does not say "missing" while the bike is on its way**

*As a* visitor, *I want* a bike's page to show the bike as soon as it arrives and not say it is missing first, *so that* a link to a bike never greets me with "That bike is no longer in the shop."

Behaviour:

1. From the moment the bike's page opens until the shop has answered for the bike in the address, the place where the bike is drawn is empty — no text, no spinner and no new screen; the header and footer stay.
2. When the shop answers with the bike, the page is drawn exactly as SITE-005 describes.
3. When the shop answers that no bike has that address, the page shows the line and button of SITE-005 Behaviour 12.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | /shop/stance-0 opened by typing the address | The page loads | At no moment before the bike appears does "That bike is no longer in the shop." show |
| AC-2 | The shop's grid | A bike card is pressed | The same holds on arrival at the bike's page |
| AC-3 | /shop/no-such-bike | The shop's answer arrives | "That bike is no longer in the shop." and "See all bikes" are shown |
| AC-4 | A bike that exists | The page has loaded | It is drawn identically to SITE-005 |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not found | The shop answers that the bike does not exist | AC-3 |
| Upstream failure | The shop cannot be reached | Not decided by the design: the page stays empty where the bike would be; see `docs/data/SRS.md` §7, open question 1 |

**Data touched** — one bike by its address.

## 5. Screens

The design is `code/frontend/src` as approved at http://localhost:9400. Rows list only the states the design draws.

| Screen | Section in the design | Functions it serves | States that must exist |
|---|---|---|---|
| Header, mobile menu, search field and panel | `SiteHeader`, `MobileMenu`, `MenuToggle` | SITE-001, SITE-007 | default, menu open, search open |
| Footer | `SiteFooter` | SITE-001 | default |
| Contact buttons | `ContactDock` | SITE-001 | default (drawn only where a number is given) |
| Home `/` | `pages/Home.tsx` | SITE-002, SITE-004, SITE-006, SITE-015 | default |
| Shop `/shop` | `pages/Shop.tsx`, `FilterRail`, `ProductCard` | SITE-003, SITE-004, SITE-006 | default, loading, no match |
| Bike `/shop/:id` | `pages/Product.tsx`, `ProductGallery`, `SpecList`, `ProductTabs` | SITE-004, SITE-005, SITE-006, SITE-013, SITE-016 | default, not found |
| Bag `/bag` | `pages/Bag.tsx` | SITE-007 | filled, empty |
| Bag drawer | `CartDrawer` | SITE-007 | filled, empty |
| Checkout `/checkout` | `pages/Checkout.tsx`, `OrderForm` | SITE-008 | default, sending, failed, not connected, empty bag |
| Order received `/checkout/done` | `pages/CheckoutDone.tsx` | SITE-009 | order, no order |
| Gallery `/gallery` | `pages/Gallery.tsx` | SITE-010 | default |
| About `/about` | `pages/About.tsx`, `ReviewCard` | SITE-011 | default, all reviews shown |
| Contact `/contact` | `pages/Contact.tsx`, `ContactForm` | SITE-012 | idle, sending, sent, failed, not connected |
| Ask band (on a bike's page) | `AskForm` | SITE-013 | idle, sending, sent, failed, not connected |
| Search `/search` | `pages/Search.tsx` | SITE-014 | nothing typed, results, nothing found |

## 6. Non-functional requirements

| Area | Requirement |
|---|---|
| Responsive | Every storefront page works from 360 px wide up with no horizontal page scroll; the layout switches at 1024 px (SITE-015) |
| Accessibility | Every image has `alt` text naming the bike (and the view for a close-up) or is decorative with empty `alt`; an icon-only control carries an `aria-label`; every form control is inside a `<label>`; each page has exactly one `h1` and headings do not skip a level; Escape closes the menu, search panel and bag drawer; the browser's focus ring is never removed from a link or button without a replacement; text contrast meets the ratios in `design/design-system.md` §1.1 (the field edge at 2.9:1 is a recorded known deviation) |
| Localisation | Copy is in English. Prices are written as `1.200.000 ₫` and dates as `16/10/2026` (`14:30` for times), through the site's formatters |
| Privacy | The message and ask forms send the visitor's name, phone or email and message to the owner's Agent Team inbox and keep nothing on the storefront. The bag, the hearts and the last order live only in the visitor's own browser |
| Motion | The owner chose "lively" motion; with the visitor's "reduce motion" setting on, nothing moves and nothing is hidden waiting to move |

## 7. Dependencies and assumptions

- **Depends on:** the `data` module, for every record the pages read once they come from the database; the `orders` module, for what "Place order" does; the `admin` module, for the owner's edits to bikes and stock; Agent Team's inbox, through the key in `src/site.json`, for the contact, ask and order messages.
- **Assumption:** the bikes' prices in đồng, stock figures, ratings, review counts, the team, the workshop prices and the shop's delivery terms are samples the team made up and marked `_invented`; the footer says "Prices in đồng are samples until the shop confirms them." The owner confirms or edits them in the Studio or the management area before launch. If any is wrong, the page shows it as the record says — nothing here changes.
- **Assumption:** the shop's phone, Zalo, email, address and hours are empty in the shipped design and are drawn only where given.

| Open question | Proposed default | Who decides |
|---|---|---|
| 1. The brief asks for "a way to call or message". The message form is built; "call" needs a phone number and the owner has given none | Ship with the message form; the owner enters the phone and Zalo number in the Studio (contact words) and the call button, the footer and the contact box appear at once | Stakeholder |
| 2. A bike the owner sets to draft no longer appears in the shop, on Home or in search, but its page still opens by its address and can still be put in the bag (the order is then refused with a sentence, ORD-004) | Leave as designed | Stakeholder |

## 8. Traceability

| Plan item | Requirement ids | Test cases |
|---|---|---|
| The site as designed | SITE-001, SITE-002, SITE-003, SITE-004, SITE-005, SITE-006, SITE-007, SITE-008, SITE-009, SITE-010, SITE-011, SITE-012, SITE-013, SITE-014, SITE-015 | `test-cases/the-site-as-designed.md` |
| Bike page waits for its record | SITE-016 | `test-cases/bike-page-waits-for-its-record.md` |
