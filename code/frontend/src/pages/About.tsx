import { useState } from "react";
import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import { formatPrice } from "../lib/format";
import type { Faq, Member, Review, Service, ShopInfo } from "../lib/shop";
import Grid from "../components/Grid";
import Icon from "../components/Icon";
import ReviewCard from "../components/ReviewCard";
import { TINTS } from "../components/TintCard";
import Weights from "../components/Weights";

/**
 * WHO IS BEHIND IT. Opens on the workshop photograph, large, with the story
 * beside it and the shop's numbers set as the template's proof block. Then
 * the people (names and what they do, on tinted cards — no strangers' faces),
 * the workshop's prices as a table with the prices lined up at the right,
 * what riders say, and the questions asked before buying.
 */
export default function About() {
  usePageMeta("pages.about.title", "pages.about.description");
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const team = useCollection<Member>("team", { pageSize: 12 }).items;
  const services = useCollection<Service>("services", { pageSize: 24 }).items;
  const reviews = useCollection<Review>("reviews", { pageSize: 6 });
  const faqs = useCollection<Faq>("faqs", { pageSize: 24 }).items;
  const photo = useContent("pages.about.image");
  const [more, setMore] = useState(false);
  const allReviews = useCollection<Review>("reviews", { pageSize: 24 }).items;
  const shown = more ? allReviews : reviews.items;
  return (
    <div className="pb-8 pt-10">
      <section className="grid gap-10 lg:grid-cols-12 lg:items-center lg:gap-14">
        <div className="lg:col-span-5">
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.about.heading" /></h1>
          <T k="pages.about.lead" as="p" className="mt-5 max-w-[48ch] leading-relaxed text-ink" />
          <T k="pages.about.story" as="p" className="mt-4 max-w-[48ch] text-sm leading-relaxed text-ink-soft" />
          {shop && (
            <div className="mt-8 flex items-start gap-5">
              <span className="mt-4 h-px w-16 shrink-0 bg-ink" />
              <dl className="grid grid-cols-3 gap-6">
                <div><dt className="text-xs text-ink-soft"><T k="pages.about.since" /></dt><dd className="mt-1 text-2xl font-bold tabular-nums text-ink">{shop.founded}</dd></div>
                <div><dt className="text-xs text-ink-soft"><T k="pages.about.built" /></dt><dd className="mt-1 text-2xl font-bold tabular-nums text-ink" data-count="">{shop.bikes_built}</dd></div>
                <div><dt className="text-xs text-ink-soft"><T k="pages.about.rated" /></dt><dd className="mt-1 text-2xl font-bold tabular-nums text-ink">{shop.rating}</dd></div>
              </dl>
            </div>
          )}
        </div>
        {photo && (
          <div className="overflow-hidden rounded lg:col-span-7">
            <Picture src={photo} alt="" width={1280} height={854} loading="eager" motion="parallax" className="aspect-[4/3] w-full object-cover" />
          </div>
        )}
      </section>

      {team.length > 0 && (
        <section className="pt-24">
          <div className="mb-10 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="pages.about.team" /></h2>
            <T k="pages.about.team.line" as="p" className="max-w-[44ch] text-sm text-ink-soft" />
          </div>
          <Grid max={4}>
            {team.map((m, i) => (
              <article key={m.id} data-record={`team:${m.id}`} className={`flex h-full flex-col gap-4 rounded p-6 ${TINTS[i % TINTS.length]}`}>
                <span className="self-start rounded-full bg-ground px-3 py-1 text-xs text-ink">{m.role}</span>
                <p className="text-[1.75rem] font-bold leading-none tracking-tight text-ink">{m.name}</p>
                <p className="flex-1 text-sm leading-relaxed text-ink">{m.line}</p>
                {m.rides && <p className="flex items-center gap-2 border-t border-line pt-3 text-xs text-ink-soft"><Icon name="people" size={14} /><T k="pages.about.rides" /> {m.rides}</p>}
              </article>
            ))}
          </Grid>
        </section>
      )}

      {services.length > 0 && (
        <section id="workshop" className="grid scroll-mt-24 gap-10 pt-24 lg:grid-cols-12">
          <div className="lg:col-span-4">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="pages.about.workshop" /></h2>
            <T k="pages.about.workshop.line" as="p" className="mt-4 max-w-[40ch] text-sm leading-relaxed text-ink-soft" />
            {shop && (
              <p className="mt-6 inline-flex items-center gap-2 rounded-full bg-[var(--tint-2)] px-4 py-2 text-sm text-ink">
                <Icon name="gift" size={16} /><T k="pages.about.free_a" /> <span className="tabular-nums">{shop.free_service_months}</span> <T k="pages.about.free_b" />
              </p>
            )}
          </div>
          <table className="w-full text-sm lg:col-span-8">
            <thead>
              <tr className="border-b border-ink text-left text-xs uppercase tracking-wider text-ink-soft">
                <th className="pb-3 font-medium"><T k="pages.about.col_job" /></th>
                <th className="hidden pb-3 text-right font-medium sm:table-cell"><T k="pages.about.col_time" /></th>
                <th className="pb-3 text-right font-medium"><T k="pages.about.col_price" /></th>
              </tr>
            </thead>
            <tbody>
              {services.map((s) => (
                <tr key={s.id} data-record={`services:${s.id}`} className="border-b border-line align-top">
                  <td className="py-4 pr-4"><span className="block font-semibold text-ink">{s.name}</span><span className="mt-1 block text-xs text-ink-soft">{s.line}</span></td>
                  <td className="hidden py-4 text-right text-ink-soft sm:table-cell">{s.time}</td>
                  <td className="py-4 text-right font-semibold tabular-nums text-ink">{s.price > 0 ? formatPrice(s.price) : <T k="pages.about.included" />}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}

      {shown.length > 0 && (
        <section id="reviews" className="scroll-mt-24 pt-24">
          <div className="mb-10 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.reviews" /></h2>
            {shop && (
              <p className="text-sm text-ink-soft"><span className="text-2xl font-bold tabular-nums text-ink">{shop.rating}</span> <T k="home.reviews.from" /> <span className="font-semibold tabular-nums text-ink">{shop.review_count}</span> <T k="home.reviews.ratings" /></p>
            )}
          </div>
          <Grid max={3} phone={1}>{shown.map((r) => <ReviewCard key={r.id} review={r} />)}</Grid>
          {!more && reviews.total > reviews.items.length && (
            <button type="button" onClick={() => setMore(true)} data-testid="reviews-more"
              className="mt-10 rounded-full border border-ink px-7 py-3 text-sm font-medium text-ink hover:bg-accent hover:text-accent-ink">
              <T k="pages.about.more_reviews" /> <span className="tabular-nums text-ink-soft">({reviews.total})</span>
            </button>
          )}
        </section>
      )}

      {faqs.length > 0 && (
        <section className="grid gap-10 pt-24 lg:grid-cols-12">
          <div className="lg:col-span-4">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="pages.about.faq" /></h2>
            <Link to="/contact" className="mt-6 inline-flex items-center gap-2 rounded-full bg-accent px-6 py-3 text-sm font-medium text-accent-ink"><T k="pages.about.ask" /><Icon name="arrow" size={16} /></Link>
          </div>
          <div className="divide-y divide-line border-y border-line lg:col-span-8">
            {faqs.map((f) => (
              <details key={f.id} data-record={`faqs:${f.id}`} className="group py-5">
                <summary className="flex cursor-pointer list-none items-center justify-between gap-6 font-semibold text-ink">
                  {f.q}
                  <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full border border-line group-open:rotate-45"><Icon name="plus" size={15} /></span>
                </summary>
                <p className="mt-3 max-w-[65ch] text-sm leading-relaxed text-ink-soft">{f.a}</p>
              </details>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
