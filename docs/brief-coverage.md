# Brief coverage — Đèo Bikes

Every line the owner chose on the brief form, quoted as the brief writes it, with the requirement ids that
meet it. Ids are in `docs/site/SRS.md` (`SITE-`), `docs/data/SRS.md` (`DATA-`), `docs/orders/SRS.md` (`ORD-`)
and `docs/admin/SRS.md` (`ADMIN-`). The brief carries no "WHAT THE FEATURES NEED" section and no documents in
`data/owner/`, so there is nothing to cover under those two headings.

## WHAT IT MUST DO

| Brief line | Requirement ids | How it is met |
|---|---|---|
| show what is offered — the menu, the products, the services — with prices | SITE-002, SITE-003, SITE-005, SITE-011, SITE-014, DATA-001, DATA-008, ORD-002, ADMIN-005, ADMIN-010, ADMIN-011 | The bikes with prices on Home, in the shop and on each bike's page; the workshop jobs with prices on About and in search; the prices live in the database and are changed by the owner in the management area; an order is priced from them |
| a gallery of work or pictures, shown large | SITE-010, SITE-005, DATA-001 | The gallery page with every photograph large, and each bike's page with a large main picture and two close-ups |
| a cart and online ordering | SITE-007, SITE-008, SITE-009, ORD-001, ORD-002, ORD-003, ORD-004, ORD-005, ORD-006, ORD-007, ORD-008 | The bag and bag drawer, the three-step checkout, the confirmation page, and an order the server stores, prices, numbers and holds the bike for |
| search across the site, with filters | SITE-001, SITE-003, SITE-014, DATA-002, DATA-003, DATA-004, DATA-005, DATA-007, DATA-008 | The header search, the search page across bikes, riding types, workshop jobs, questions and gallery, and the shop's filters by riding type, frame size, wheel size, brand and price |
| a way to call or message | SITE-001, SITE-012, SITE-013 | A message form on the contact page, an ask-the-shop-to-call form on every bike's page, and call and Zalo buttons on a phone. **The call button needs a phone number the owner has not yet given** — see `docs/site/SRS.md` §7, open question 1 |
| who is behind it: the story, the team, what customers say | SITE-011, SITE-002, SITE-005 | The About page with the story, the four team members and the reviews; reviews also on Home and on each bike's page |
| Stock availability status (in the owner's words) | SITE-004, SITE-003, SITE-005, ORD-004, ORD-005, ADMIN-005, ADMIN-007, ADMIN-009, ADMIN-010, ADMIN-011 | "In stock", "Out of stock" and "· N left" on every bike, in words the owner edits; an order lowers the stock; the owner keeps stock true in the management area and it shows at once |

## HOW CUSTOMERS GET IN TOUCH

| Brief line | Requirement ids | How it is met |
|---|---|---|
| a message form on the site — name, phone or email, message; the owner reads what is sent in Agent Team | SITE-012, SITE-013, ORD-008 | The contact page's form (name, phone or email, message) and the bike page's ask form send to the owner's inbox in Agent Team; an order also reaches Agent Team as one message |

## Also traced: the owner's description

The description the owner wrote, line by line, so nothing in it is unplanned.

| Description line | Requirement ids |
|---|---|
| Khách vào trang chủ sẽ thấy ngay các dòng xe nổi bật và có thể xem toàn bộ xe đang bán. | SITE-002, SITE-003 |
| Họ lọc danh sách theo loại xe, cỡ khung, cỡ bánh, thương hiệu và mức giá, rồi mở trang chi tiết của từng chiếc. | SITE-003, SITE-005, DATA-003, DATA-008 |
| Mỗi trang có nhiều ảnh lớn, giá bán, thông số kỹ thuật như khung, phuộc, bộ truyền động, phanh và trọng lượng, cùng tình trạng còn hàng hay hết hàng. | SITE-005, SITE-004 |
| Khi đã chọn được xe, khách đặt mua ngay trên trang hoặc để lại thông tin để cửa hàng liên hệ tư vấn và chốt đơn. | SITE-005, SITE-007, SITE-008, SITE-009, SITE-013, ORD-001, ORD-008 |
| Trang cần hiển thị tốt trên điện thoại, vì phần lớn khách sẽ xem xe bằng điện thoại, và phải có mục giới thiệu cửa hàng cùng thông tin liên hệ rõ ràng. | SITE-015, SITE-011, SITE-012, SITE-001 |
| Giao diện nên gọn, nhiều ảnh, mang cảm giác khỏe khoắn, năng động và đáng tin cậy. | `design/design-system.md`; SITE-002, SITE-010, SITE-015 (the approved design, not a requirement of its own) |
