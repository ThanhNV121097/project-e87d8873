import { Link } from "react-router-dom";
import { T } from "../editable";
import { useCollection } from "../data";
import type { Department } from "../lib/shop";
import PhotoCredits from "./PhotoCredits";

/**
 * THE FOOTER: white, under a hairline that stops at the measure like the
 * header's — the name in bold and one line at the left, then the departments
 * and the ways to reach the shop. The stock photographs' credits sit at its
 * foot, small: their licences ask for them.
 */
export default function SiteFooter() {
  const departments = useCollection<Department>("categories", { pageSize: 24 }).items;
  return (
    <footer className="mx-auto mt-20 max-w-page px-[var(--gutter)] pb-10 text-ink">
      <div className="flex flex-col gap-10 border-t border-line pt-12 text-sm lg:flex-row lg:gap-16">
        <div className="lg:w-5/12">
          <T k="site.name" as="p" className="text-2xl font-bold tracking-tight" />
          <T k="footer.line" as="p" className="mt-3 max-w-[38ch] leading-relaxed text-ink-soft" />
        </div>
        {departments.length > 0 && (
          <nav aria-label="Departments" className="lg:w-4/12">
            <T k="footer.departments" as="p" className="mb-3 font-semibold" />
            <ul className="grid grid-cols-2 gap-x-6 gap-y-2">
              {departments.map((d) => (
                <li key={d.id}><Link to={`/shop?category=${d.id}`} className="text-ink-soft hover:text-ink">{d.name}</Link></li>
              ))}
            </ul>
          </nav>
        )}
        <nav aria-label="Help" className="lg:w-3/12">
          <T k="footer.help" as="p" className="mb-3 font-semibold" />
          <ul className="flex flex-col gap-2">
            <li><Link to="/contact" className="text-ink-soft hover:text-ink"><T k="footer.contact" /></Link></li>
            <li><Link to="/bag" className="text-ink-soft hover:text-ink"><T k="footer.bag" /></Link></li>
            <li><Link to="/admin" className="text-ink-soft hover:text-ink"><T k="footer.admin" /></Link></li>
          </ul>
        </nav>
      </div>
      <PhotoCredits className="mt-10 text-xs text-ink-soft" />
    </footer>
  );
}
