import { Link } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { useCollection } from "../data";
import type { Department, NavItem } from "../lib/shop";
import Icon from "./Icon";
import Logo from "./Logo";
import PhotoCredits from "./PhotoCredits";

/**
 * THE FOOTER: white, under a hairline that stops at the measure like the
 * header's — the mark and name with one line at the left, then the riding
 * types, the pages, and how to reach the shop. A way to reach the shop is
 * drawn only where the owner has given it; the message form always is. At
 * its foot the line saying the shop is independent of the brands it sells,
 * and the stock photographs' credits, small.
 */
export default function SiteFooter() {
  const types = useCollection<Department>("categories", { pageSize: 24 }).items;
  const links = useList<NavItem>("nav.links");
  const phone = useContent("contact.phone");
  const email = useContent("contact.email");
  const address = useContent("contact.address");
  const hours = useContent("contact.hours");
  return (
    <footer className="mx-auto mt-24 max-w-page px-[var(--gutter)] pb-[calc(var(--dock-space)+1rem)] text-ink lg:pb-10">
      <div className="grid gap-10 border-t border-line pt-12 text-sm sm:grid-cols-2 lg:grid-cols-12 lg:gap-8">
        <div className="lg:col-span-4">
          <Logo />
          <T k="footer.line" as="p" className="mt-4 max-w-[34ch] leading-relaxed text-ink-soft" />
        </div>
        {types.length > 0 && (
          <nav aria-label="Riding types" className="lg:col-span-2">
            <T k="footer.types" as="p" className="mb-3 font-semibold" />
            <ul className="flex flex-col gap-2">
              {types.map((d) => (
                <li key={d.id}><Link to={`/shop?category=${d.id}`} className="text-ink-soft hover:text-ink">{d.name}</Link></li>
              ))}
            </ul>
          </nav>
        )}
        <nav aria-label="Pages" className="lg:col-span-2">
          <T k="footer.pages" as="p" className="mb-3 font-semibold" />
          <ul className="flex flex-col gap-2">
            {links.map((l, i) => (
              <li key={l.href}><Link to={l.href} className="text-ink-soft hover:text-ink"><T k={`nav.links.${i}.label`} /></Link></li>
            ))}
            <li><Link to="/bag" className="text-ink-soft hover:text-ink"><T k="footer.bag" /></Link></li>
          </ul>
        </nav>
        <div className="lg:col-span-4">
          <T k="footer.reach" as="p" className="mb-3 font-semibold" />
          <ul className="flex flex-col gap-2.5 text-ink-soft">
            {phone && <li className="flex items-center gap-2.5"><Icon name="phone" size={16} /><a href={`tel:${phone}`} className="hover:text-ink">{phone}</a></li>}
            {email && <li className="flex items-center gap-2.5"><Icon name="mail" size={16} /><a href={`mailto:${email}`} className="hover:text-ink">{email}</a></li>}
            {address && <li className="flex items-start gap-2.5"><Icon name="address" size={16} /><span>{address}</span></li>}
            {hours && <li className="flex items-start gap-2.5"><Icon name="hours" size={16} /><span>{hours}</span></li>}
            <li className="flex items-center gap-2.5"><Icon name="chat" size={16} /><Link to="/contact" className="hover:text-ink"><T k="footer.message" /></Link></li>
            <li className="flex items-center gap-2.5"><Icon name="lock" size={16} /><Link to="/admin" className="hover:text-ink"><T k="footer.admin" /></Link></li>
          </ul>
        </div>
      </div>
      <div className="mt-10 flex flex-col gap-3 border-t border-line pt-6 text-xs text-ink-soft lg:flex-row lg:justify-between">
        <T k="footer.independent" as="p" className="max-w-[70ch]" />
        <PhotoCredits className="max-w-[70ch]" />
      </div>
    </footer>
  );
}
