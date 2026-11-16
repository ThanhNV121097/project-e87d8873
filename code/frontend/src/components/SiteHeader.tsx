import { useState, type FormEvent } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { useCart } from "../lib/cart";
import type { NavItem } from "../lib/shop";
import CartDrawer from "./CartDrawer";
import Icon from "./Icon";
import Logo from "./Logo";
import MenuToggle from "./MenuToggle";
import MobileMenu from "./MobileMenu";

/**
 * THE HEADER: the pass mark and the name at the left, the links in the
 * middle, and at the right search, the owner's sign-in and the bag — one
 * quiet row BETWEEN TWO HAIRLINES that stop at the page's measure. Nothing
 * in it is coloured: the colour on this site belongs to the cards.
 *
 * Search opens a field under the row and sends what is typed to /search,
 * which looks through the bikes, the riding types, the workshop's jobs and
 * prices, the buying questions and the gallery.
 */
export default function SiteHeader() {
  const searchLabel = useContent("header.search");
  const searchHint = useContent("header.search_hint");
  const bagLabel = useContent("header.bag");
  const accountLabel = useContent("header.account");
  const links = useList<NavItem>("nav.links");
  const { count } = useCart();
  const [menu, setMenu] = useState(false);
  const [bag, setBag] = useState(false);
  const [search, setSearch] = useState(false);
  const [q, setQ] = useState("");
  const go = useNavigate();
  const submit = (e: FormEvent) => {
    e.preventDefault();
    setSearch(false);
    setMenu(false);
    go(q.trim() ? `/search?q=${encodeURIComponent(q.trim())}` : "/search");
  };
  const iconButton = "grid h-10 w-10 place-items-center rounded-full hover:bg-wash";
  return (
    <header className="mx-auto max-w-page px-[var(--gutter)] pt-4 text-ink">
      <div className="flex h-[var(--header-h)] items-center gap-6 border-y border-line">
        <Link to="/" className="shrink-0"><Logo /></Link>
        <nav aria-label="Main" className="mx-auto hidden items-center gap-9 text-sm lg:flex">
          {links.map((l, i) => (
            <NavLink key={l.href} to={l.href} end={l.href === "/"}
              className={({ isActive }) => `underline-offset-4 hover:underline ${isActive ? "font-semibold" : ""}`}>
              <T k={`nav.links.${i}.label`} />
            </NavLink>
          ))}
        </nav>
        <div className="ml-auto flex shrink-0 items-center gap-1 lg:ml-0">
          <button type="button" onClick={() => { setMenu(false); setSearch((s) => !s); }} aria-label={searchLabel} aria-expanded={search}
            data-testid="open-search" className={iconButton}>
            <Icon name={search ? "close" : "search"} size={19} />
          </button>
          <Link to="/admin" aria-label={accountLabel} className={`hidden lg:grid ${iconButton}`}><Icon name="account" size={19} /></Link>
          <button type="button" onClick={() => setBag(true)} data-testid="open-bag" aria-label={bagLabel} className={`relative ${iconButton}`}>
            <Icon name="bag" size={19} />
            {count > 0 && (
              <span className="absolute right-0.5 top-0.5 grid h-4 min-w-4 place-items-center rounded-full bg-accent px-1 text-[0.625rem] tabular-nums text-accent-ink">{count}</span>
            )}
          </button>
          <MenuToggle open={menu} onToggle={() => { setSearch(false); setMenu((o) => !o); }} className="lg:hidden" />
        </div>
      </div>

      {search && (
        <form role="search" onSubmit={submit} className="hidden items-center gap-3 border-b border-line py-4 lg:flex">
          <Icon name="search" size={20} className="text-ink-soft" />
          <input autoFocus value={q} onChange={(e) => setQ(e.target.value)} aria-label={searchLabel} placeholder={searchHint}
            data-testid="search-input"
            className="min-w-0 flex-1 appearance-none bg-transparent text-lg text-ink outline-none placeholder:text-ink-soft" />
          <button type="submit" className="whitespace-nowrap rounded-full bg-accent px-5 py-2 text-sm font-medium text-accent-ink"><T k="header.search_go" /></button>
        </form>
      )}

      {/* Below 1024px search is a solid panel over the screen, as the menu is. */}
      <MobileMenu open={search} onClose={() => setSearch(false)}>
        <form role="search" onSubmit={submit} className="flex flex-col gap-4">
          <T k="header.search" as="p" className="text-3xl font-light tracking-tight text-ink" />
          <input value={q} onChange={(e) => setQ(e.target.value)} aria-label={searchLabel} placeholder={searchHint}
            className="w-full appearance-none rounded-[var(--radius-sm)] border border-field bg-ground px-4 py-3 text-ink placeholder:text-ink-soft" />
          <button type="submit" className="inline-flex items-center justify-center gap-2 self-start whitespace-nowrap rounded-full bg-accent px-6 py-3 text-sm font-medium text-accent-ink">
            <Icon name="search" size={16} /><T k="header.search_go" />
          </button>
        </form>
      </MobileMenu>

      <MobileMenu open={menu} onClose={() => setMenu(false)}>
        <nav className="flex flex-col gap-5 text-3xl font-light tracking-tight text-ink" aria-label="Main">
          {links.map((l, i) => (
            <Link key={l.href} to={l.href} onClick={() => setMenu(false)}><T k={`nav.links.${i}.label`} /></Link>
          ))}
          <Link to="/bag" onClick={() => setMenu(false)}><T k="header.bag" /></Link>
          <Link to="/admin" onClick={() => setMenu(false)} className="mt-4 text-base font-normal text-ink-soft"><T k="header.account" /></Link>
        </nav>
      </MobileMenu>

      <CartDrawer open={bag} onClose={() => setBag(false)} checkout="/checkout"
        title={<T k="pages.bag.heading_plain" />} empty={<T k="pages.bag.empty" />}
        pay={<T k="pages.bag.drawer_pay" />} totalLabel={<T k="pages.bag.total" />} />
    </header>
  );
}
