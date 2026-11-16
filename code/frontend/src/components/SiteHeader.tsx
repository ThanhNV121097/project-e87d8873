import { useState } from "react";
import { Link } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { useCart } from "../lib/cart";
import type { NavItem } from "../lib/shop";
import CartDrawer from "./CartDrawer";
import Icon from "./Icon";
import MenuToggle from "./MenuToggle";
import MobileMenu from "./MobileMenu";

/**
 * THE HEADER: the name in bold at the left, the links in the middle, and at
 * the right search, the sign-in with its word, and the bag — one quiet row
 * BETWEEN TWO HAIRLINES that stop at the page's measure. Nothing in it is
 * coloured: the colour on this page belongs to the cards.
 */
export default function SiteHeader() {
  const name = useContent("site.name");
  const searchLabel = useContent("header.search");
  const bagLabel = useContent("header.bag");
  const links = useList<NavItem>("nav.links");
  const { count } = useCart();
  const [menu, setMenu] = useState(false);
  const [bag, setBag] = useState(false);
  return (
    <header className="mx-auto max-w-page px-[var(--gutter)] pt-4 text-ink">
      <div className="flex h-[var(--header-h)] items-center gap-8 border-y border-line">
        <Link to="/" aria-label={name} className="shrink-0 text-xl font-bold tracking-tight"><T k="site.name" /></Link>
        <nav aria-label="Main" className="mx-auto hidden items-center gap-9 text-sm lg:flex">
          {links.map((l, i) => (
            <Link key={l.href} to={l.href} className="underline-offset-4 hover:underline"><T k={`nav.links.${i}.label`} /></Link>
          ))}
        </nav>
        <div className="ml-auto flex shrink-0 items-center gap-5 text-sm lg:ml-0">
          <Link to="/shop" aria-label={searchLabel} className="hidden lg:block"><Icon name="search" size={19} /></Link>
          <Link to="/admin" className="hidden items-center gap-1.5 lg:inline-flex"><Icon name="account" size={18} /><T k="header.account" /></Link>
          <button type="button" onClick={() => setBag(true)} data-testid="open-bag" aria-label={bagLabel} className="relative">
            <Icon name="bag" size={19} />
            {count > 0 && (
              <span className="absolute -right-2 -top-2 grid h-4 min-w-4 place-items-center rounded-full bg-accent px-1 text-[0.625rem] tabular-nums text-accent-ink">{count}</span>
            )}
          </button>
          <MenuToggle open={menu} onToggle={() => setMenu((o) => !o)} className="lg:hidden" />
        </div>
      </div>

      <MobileMenu open={menu} onClose={() => setMenu(false)}>
        <nav className="flex flex-col gap-5 text-3xl font-light tracking-tight text-ink" aria-label="Main">
          {links.map((l, i) => (
            <Link key={l.href} to={l.href} onClick={() => setMenu(false)}><T k={`nav.links.${i}.label`} /></Link>
          ))}
          <Link to="/admin" onClick={() => setMenu(false)} className="text-base font-normal text-ink-soft"><T k="header.account" /></Link>
        </nav>
      </MobileMenu>

      <CartDrawer open={bag} onClose={() => setBag(false)} checkout="/checkout"
        title={<T k="pages.bag.heading" />} empty={<T k="pages.bag.empty" />}
        pay={<T k="pages.bag.drawer_pay" />} totalLabel={<T k="pages.bag.total" />} />
    </header>
  );
}
