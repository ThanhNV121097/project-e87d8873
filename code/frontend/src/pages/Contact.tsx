import { T, useContent } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import type { ShopInfo } from "../lib/shop";
import ContactForm from "../components/ContactForm";
import Icon, { type IconName } from "../components/Icon";
import MapEmbed from "../components/MapEmbed";
import { TINTS } from "../components/TintCard";
import Weights from "../components/Weights";

const REASONS: IconName[] = ["people", "date", "settings"];

/**
 * HOW TO REACH THE SHOP. The form at the left — name, phone or email, the
 * bike and the rider's height, the message — and at the right a photograph
 * and the ways to reach the shop, each with its icon, drawn only where the
 * owner has given them. Under both, the three things people write about, as
 * the template's feature row.
 */
export default function Contact() {
  usePageMeta("pages.contact.title", "pages.contact.description");
  const phone = useContent("contact.phone");
  const email = useContent("contact.email");
  const address = useContent("contact.address");
  const hours = useContent("contact.hours");
  const photo = useContent("pages.contact.image");
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const details: [IconName, string, string | undefined][] = [
    ["phone", phone, phone ? `tel:${phone}` : undefined], ["mail", email, email ? `mailto:${email}` : undefined],
    ["address", address, undefined], ["hours", hours, undefined],
  ];
  const given = details.filter(([, v]) => Boolean(v));
  return (
    <div className="pb-8 pt-10">
      <section className="grid gap-12 lg:grid-cols-12 lg:gap-14">
        <div className="lg:col-span-7">
          <h1 data-testid="contact-heading" className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.contact.heading" /></h1>
          <T k="pages.contact.lead" as="p" className="mt-4 max-w-[56ch] text-sm leading-relaxed text-ink-soft" />
          {shop && (
            <p className="mt-5 inline-flex items-center gap-2 rounded-full bg-[var(--tint-2)] px-4 py-2 text-sm text-ink">
              <Icon name="hours" size={16} /><T k="pages.contact.reply_a" /> <span className="tabular-nums">{shop.reply_hours}</span> <T k="pages.contact.reply_b" />
            </p>
          )}
          <ContactForm className="mt-10 max-w-[40rem]" fields={[{ key: "bike" }, { key: "height", type: "number" }]} />
        </div>
        {/* On a phone the photograph leads, edge to edge, above the heading. */}
        <aside className="order-first flex flex-col gap-6 lg:order-none lg:col-span-5">
          {photo && (
            <div className="-mx-[var(--gutter)] overflow-hidden sm:mx-0 sm:rounded lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:rounded-r-none">
              <Picture src={photo} alt="" width={853} height={1280} loading="eager" className="aspect-[3/2] w-full object-cover lg:aspect-[4/5]" />
            </div>
          )}
          {given.length > 0 && (
            <ul className="flex flex-col gap-4 rounded border border-line p-6 text-sm text-ink">
              {given.map(([icon, value, href]) => (
                <li key={icon} className="flex items-start gap-3">
                  <Icon name={icon} size={18} className="text-ink-soft" />
                  {href ? <a href={href} className="font-medium underline-offset-4 hover:underline">{value}</a> : <span>{value}</span>}
                </li>
              ))}
            </ul>
          )}
        </aside>
      </section>

      <section className="grid gap-4 pt-24 md:grid-cols-3">
        {REASONS.map((icon, i) => (
          <div key={icon} className={`flex flex-col gap-3 rounded p-6 ${TINTS[i]}`}>
            <span className="grid h-11 w-11 place-items-center rounded-full bg-ground text-ink"><Icon name={icon} size={19} /></span>
            <T k={`pages.contact.reasons.${i}.bold`} as="p" className="font-semibold text-ink" />
            <T k={`pages.contact.reasons.${i}.light`} as="p" className="text-sm leading-relaxed text-ink-soft" />
          </div>
        ))}
      </section>

      {address && <section className="pt-24"><MapEmbed className="h-80 overflow-hidden rounded" /></section>}
    </div>
  );
}
