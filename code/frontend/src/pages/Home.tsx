import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import { SplitHeading } from "../motion";
import type { Department, Product, Review, ShopInfo } from "../lib/shop";
import Grid from "../components/Grid";
import Icon, { type IconName } from "../components/Icon";
import ProductCard from "../components/ProductCard";
import ProductImage from "../components/ProductImage";
import ReviewCard from "../components/ReviewCard";
import DeliveryTerms from "../components/DeliveryTerms";
import TintCard from "../components/TintCard";
import Weights from "../components/Weights";

const PROMISE_ICONS: IconName[] = ["shipping", "secure", "chat"];

/**
 * THE HOME PAGE, as the owner's template has it. No photograph opens it: a
 * heading over three lines whose words change weight, the dark pill ON its
 * last line, and a short block of proof at its right — then three tall
 * tinted cards with the bikes the shop leads with. The colour on this page is
 * in the cards and in one peach band, nowhere else.
 *
 * Then the riding types as grey wordmarks between two hairlines, the bike of
 * the month, a grey band of where to start, the three promises, more bikes
 * three to a row, and what riders say.
 */
export default function Home() {
  usePageMeta("pages.home.title", "pages.home.description");
  const lead = useCollection<Product>("products", { where: { status: "live", deal: true }, pageSize: 3 }).items;
  const inStock = useCollection<Product>("products", { where: { status: "live", stock: { from: 1 } }, pageSize: 3 });
  const month = useCollection<Product>("products", { where: { status: "live", month: true }, pageSize: 1 }).items[0];
  const picks = useCollection<Product>("products", { where: { status: "live", deal: false, month: undefined }, pageSize: 6 }).items;
  const types = useCollection<Department>("categories", { pageSize: 12 }).items;
  const reviews = useCollection<Review>("reviews", { pageSize: 3 }).items;
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const nameOf = (id: string) => types.find((d) => d.id === id)?.name;
  return (
    <>
      <section className="flex flex-col gap-10 pb-12 pt-12 lg:flex-row lg:items-center lg:justify-between lg:pb-10 lg:pt-12">
        <h1 className="text-[length:var(--size-h1)] leading-[1.12] tracking-[-0.01em] text-ink">
          <SplitHeading k="hero.headline.light" as="span" className="block font-light" />
          <SplitHeading k="hero.headline.bold" as="span" className="block font-bold" />
          {/* THE PILL SITS ON THE LAST LINE, beside its words. */}
          <span className="flex flex-wrap items-center gap-x-8 gap-y-4">
            <SplitHeading k="hero.headline.tail" as="span" className="font-light" />
            <Link to="/shop" data-testid="hero-cta" data-magnetic=""
              className="inline-flex items-center gap-2 whitespace-nowrap rounded-full bg-accent px-9 py-3.5 font-body text-sm font-medium tracking-normal text-accent-ink hover:opacity-90">
              <T k="hero.cta.label" /><Icon name="arrow" size={16} />
            </Link>
          </span>
        </h1>
        {/* THE PROOF: three of the shop's own bikes as overlapping rounds, how
            many are in stock — counted from the records — and two lines. */}
        {inStock.total > 0 && (
          <div className="flex items-start gap-5 lg:w-[24rem]">
            <span className="mt-14 hidden h-px w-20 shrink-0 bg-ink lg:block" />
            <div>
              <div className="flex items-center gap-3">
                <span className="flex -space-x-3">
                  {inStock.items.map((p) => (
                    <span key={p.id} className="block w-12 overflow-hidden rounded-full border border-line ring-2 ring-[var(--ground)]">
                      <ProductImage src={p.image_detail ?? p.image} alt="" aspect="1/1" eager className="rounded-none" />
                    </span>
                  ))}
                </span>
                <span className="text-sm font-bold leading-tight text-ink">
                  <span className="tabular-nums" data-count="">{inStock.total}</span>{" "}<T k="hero.proof.count" />
                </span>
              </div>
              <T k="hero.proof.line" as="p" className="mt-3 max-w-[30ch] text-sm leading-snug text-ink" />
            </div>
          </div>
        )}
      </section>

      {lead.length > 0 && (
        <Grid max={3} phone={1} className="items-stretch">
          {lead.map((p, i) => <TintCard key={p.id} product={p} tint={i} label={nameOf(p.category)} eager />)}
        </Grid>
      )}

      {/* The departments as wordmarks, between two hairlines. */}
      {types.length > 0 && (
        <nav aria-label="Departments" className="mt-20 flex flex-wrap items-center gap-x-8 gap-y-4 border-y border-line py-9 lg:justify-between lg:gap-x-10">
          {types.map((d) => (
            <Link key={d.id} to={`/shop?category=${d.id}`} data-record={`categories:${d.id}`} data-testid={`mark-${d.id}`}
              className="text-sm font-semibold uppercase tracking-[0.2em] text-ink-soft hover:text-ink lg:text-lg">{d.name}</Link>
          ))}
          <Link to="/about#workshop" data-testid="mark-workshop"
            className="text-sm font-semibold uppercase tracking-[0.2em] text-ink-soft hover:text-ink lg:text-lg"><T k="home.workshop_mark" /></Link>
        </nav>
      )}

      {month && <Month bike={month} type={nameOf(month.category)} />}

      {types.length > 0 && <Start types={types} />}

      {shop && (
        <section className="mt-24 grid gap-8 rounded bg-[var(--tint-4)] px-6 py-10 md:grid-cols-3 lg:px-12">
          {[shop.delivery_local, `${shop.warranty_years}`, `${shop.reply_hours}`].map((n, i) => (
            <div key={i} className="flex flex-col items-center text-center">
              <span className="grid h-12 w-12 place-items-center rounded-full bg-ground text-ink"><Icon name={PROMISE_ICONS[i]} size={20} /></span>
              <p className="mt-3 text-sm font-semibold text-ink">
                {i === 0 && <T k="home.promises.0.bold" />}
                {i === 1 && <><span className="tabular-nums">{n}</span>{" "}<T k="home.promises.1.bold" /></>}
                {i === 2 && <><T k="home.promises.2.bold" />{" "}<span className="tabular-nums">{n}</span>{" "}<T k="home.promises.2.unit" /></>}
              </p>
              {i === 0 ? (
                <DeliveryTerms shop={shop} className="mt-1 max-w-[34ch] text-xs leading-relaxed text-ink-soft" />
              ) : (
                <p className="mt-1 max-w-[28ch] text-xs leading-relaxed text-ink-soft"><T k={`home.promises.${i}.light`} /></p>
              )}
            </div>
          ))}
        </section>
      )}

      {picks.length > 0 && (
        <section className="pt-24">
          <div className="mb-10 text-center">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.picks" /></h2>
            <T k="home.picks.line" as="p" className="mx-auto mt-3 max-w-[40ch] text-sm leading-relaxed text-ink-soft" />
            <nav aria-label="Riding types" className="mt-6 flex flex-wrap justify-center gap-2 text-xs">
              <Link to="/shop" data-testid="picks-all" className="rounded-full bg-accent px-4 py-2 font-medium text-accent-ink"><T k="home.picks.all" /></Link>
              {types.map((d) => (
                <Link key={d.id} to={`/shop?category=${d.id}`} data-testid={`pick-chip-${d.id}`}
                  className="rounded-full border border-ink px-4 py-2 text-ink hover:bg-accent hover:text-accent-ink">{d.name}</Link>
              ))}
            </nav>
          </div>
          <Grid max={3} phone={1}>{picks.map((p) => <ProductCard key={p.id} product={p} />)}</Grid>
        </section>
      )}

      {reviews.length > 0 && (
        <section className="pt-24">
          <div className="mb-10 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.reviews" /></h2>
            {shop && (
              <p className="text-sm text-ink-soft">
                <span className="text-2xl font-bold tabular-nums text-ink">{shop.rating}</span>{" "}<T k="home.reviews.from" />{" "}
                <span className="font-semibold tabular-nums text-ink" data-count="">{shop.review_count}</span>{" "}<T k="home.reviews.ratings" />
                <Link to="/about#reviews" className="ml-4 inline-flex items-center gap-1 font-medium text-ink underline-offset-4 hover:underline"><T k="home.reviews.more" /><Icon name="next" size={15} /></Link>
              </p>
            )}
          </div>
          <Grid max={3} phone={1}>{reviews.map((r) => <ReviewCard key={r.id} review={r} />)}</Grid>
        </section>
      )}
    </>
  );
}

/**
 * THE BIKE OF THE MONTH, as the template sets its "top product": two tall
 * cards side by side at the left, the shorter a riding photograph, the taller
 * the bike on its tint; and at the right the heading set as large as the
 * page's own, a line and the way to the bike.
 */
function Month({ bike, type }: { bike: Product; type?: string }) {
  const scene = useContent("home.month.image");
  return (
    <section className="flex flex-col gap-10 pt-24 lg:flex-row lg:items-center lg:gap-16">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end lg:w-7/12 lg:gap-5">
        {/* The riding photograph runs off the left edge of the screen. */}
        {scene && (
          <div className="-mx-[var(--gutter)] overflow-hidden sm:mx-0 sm:min-w-0 sm:flex-1 sm:rounded lg:-ml-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:rounded-l-none">
            <Picture src={scene} alt="" width={853} height={1280} motion="parallax" className="aspect-[16/10] h-full w-full object-cover sm:aspect-[3/4]" />
          </div>
        )}
        <div className="sm:w-7/12 sm:shrink-0">
          <TintCard product={bike} tint={2} label={type} />
        </div>
      </div>
      <div className="lg:w-5/12">
        <h2 className="text-[length:var(--size-h1)] leading-[1.12] text-ink"><Weights k="home.month" /></h2>
        <p className="mt-5 max-w-[42ch] text-sm leading-relaxed text-ink-soft">{bike.description}</p>
        <dl className="mt-6 grid max-w-[26rem] grid-cols-2 gap-x-6 gap-y-3 border-t border-line pt-5 text-sm">
          <dt className="text-ink-soft"><T k="home.month.travel" /></dt><dd className="text-right font-medium text-ink">{bike.travel}</dd>
          <dt className="text-ink-soft"><T k="home.month.weight" /></dt><dd className="text-right font-medium text-ink">{bike.weight}</dd>
        </dl>
        <Link to={`/shop/${bike.id}`} data-testid="month-cta"
          className="mt-8 inline-flex items-center gap-2 rounded-full bg-accent px-7 py-3 text-sm font-medium text-accent-ink hover:opacity-90">
          <T k="home.month.cta" /><Icon name="arrow" size={16} />
        </Link>
      </div>
    </section>
  );
}

/**
 * Where to start: a quiet grey band — a heading and a row for each riding
 * type at the left, who it is for under its name and an arrow, a hairline
 * between; and a riding photograph at the right.
 */
function Start({ types }: { types: Department[] }) {
  const image = useContent("home.start.image");
  return (
    <section className="mt-24 flex flex-col gap-10 rounded bg-surface p-6 lg:flex-row lg:items-stretch lg:gap-14 lg:p-12">
      <div className="flex-1">
        <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.start" /></h2>
        <T k="home.start.line" as="p" className="mt-3 max-w-[44ch] text-sm leading-relaxed text-ink-soft" />
        <ul className="mt-8">
          {types.map((d) => (
            <li key={d.id} className="border-b border-line last:border-b-0">
              <Link to={`/shop?category=${d.id}`} data-testid={`start-${d.id}`} className="group flex items-center justify-between gap-6 py-5">
                <span>
                  <span className="block font-semibold text-ink underline-offset-4 group-hover:underline">{d.name}</span>
                  <span className="mt-1 block text-sm text-ink-soft">{d.line}</span>
                  <span className="mt-1 block text-xs text-ink-soft">{d.terrain}</span>
                </span>
                <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full border border-line bg-ground text-ink group-hover:bg-accent group-hover:text-accent-ink">
                  <Icon name="arrow" size={17} />
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
      {image && (
        <div className="overflow-hidden rounded lg:w-5/12">
          <Picture src={image} alt="" width={850} height={1280} motion="parallax" className="aspect-[4/5] h-full w-full object-cover" />
        </div>
      )}
    </section>
  );
}
