import { Link } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { useCollection } from "../data";
import type { Department, NavItem } from "../lib/shop";
import Icon from "./Icon";
import LogoMark from "./LogoMark";
import PhotoCredits from "./PhotoCredits";

/**
 * THE FOOTER: the site's one surface of navy — a large rounded card at the
 * page's measure, as the template's cards are, so the ink that presses every
 * button also closes every page. The mark and a line at the left, then the
 * riding types, the pages, and how to reach the shop: a way to reach it is
 * drawn only where the owner has given it; the message form always is. At
 * its foot the line saying the shop is independent of the brands it sells,
 * and the stock photographs' credits.
 */
export default function SiteFooter() {
  const types = useCollection<Department>("categories", { pageSize: 24 }).items;
  const links = useList<NavItem>("nav.links");
  const phone = useContent("contact.phone");
  const email = useContent("contact.email");
  const address = useContent("contact.address");
  const hours = useContent("contact.hours");
  const soft = "text-accent-ink/75 hover:text-accent-ink";
  return (
    <footer className="mx-auto mt-24 max-w-page px-[var(--gutter)] pb-[calc(var(--dock-space)+1rem)] lg:pb-8">
      <div className="rounded bg-accent px-6 py-12 text-accent-ink sm:px-10 lg:px-14 lg:py-14">
        <div className="grid gap-10 text-sm sm:grid-cols-2 lg:grid-cols-12 lg:gap-8">
          <div className="lg:col-span-4">
            <Link to="/" className="inline-flex items-center gap-2 text-lg tracking-tight"><LogoMark className="h-7 w-7" /><T k="site.name" /></Link>
            <T k="footer.line" as="p" className="mt-4 max-w-[34ch] leading-relaxed text-accent-ink/75" />
          </div>
          {types.length > 0 && (
            <nav aria-label="Riding types" className="lg:col-span-2">
              <T k="footer.types" as="p" className="mb-3 font-semibold" />
              <ul className="flex flex-col gap-2">
                {types.map((d) => (
                  <li key={d.id}><Link to={`/shop?category=${d.id}`} className={soft}>{d.name}</Link></li>
                ))}
              </ul>
            </nav>
          )}
          <nav aria-label="Pages" className="lg:col-span-2">
            <T k="footer.pages" as="p" className="mb-3 font-semibold" />
            <ul className="flex flex-col gap-2">
              {links.map((l, i) => (
                <li key={l.href}><Link to={l.href} className={soft}><T k={`nav.links.${i}.label`} /></Link></li>
              ))}
              <li><Link to="/bag" className={soft}><T k="footer.bag" /></Link></li>
            </ul>
          </nav>
          <div className="lg:col-span-4">
            <T k="footer.reach" as="p" className="mb-3 font-semibold" />
            <ul className="flex flex-col gap-2.5">
              {phone && <li className="flex items-center gap-2.5"><Icon name="phone" size={16} /><a href={`tel:${phone}`} className={soft}>{phone}</a></li>}
              {email && <li className="flex items-center gap-2.5"><Icon name="mail" size={16} /><a href={`mailto:${email}`} className={soft}>{email}</a></li>}
              {address && <li className="flex items-start gap-2.5"><Icon name="address" size={16} /><span className="text-accent-ink/75">{address}</span></li>}
              {hours && <li className="flex items-start gap-2.5"><Icon name="hours" size={16} /><span className="text-accent-ink/75">{hours}</span></li>}
              <li className="flex items-center gap-2.5"><Icon name="chat" size={16} /><Link to="/contact" className={soft}><T k="footer.message" /></Link></li>
              <li className="flex items-center gap-2.5"><Icon name="lock" size={16} /><Link to="/admin" className={soft}><T k="footer.admin" /></Link></li>
            </ul>
          </div>
        </div>
        <div className="mt-12 flex flex-col gap-3 border-t border-accent-ink/20 pt-6 text-xs text-accent-ink/75 lg:flex-row lg:justify-between lg:gap-10">
          <T k="footer.independent" as="p" className="max-w-[70ch]" />
          <PhotoCredits className="max-w-[70ch] text-accent-ink/80 [&_a]:text-accent-ink" />
        </div>
      </div>
    </footer>
  );
}
