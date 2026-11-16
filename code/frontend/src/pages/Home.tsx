import { Link } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import type { Department, Feature, Product } from "../lib/shop";
import Grid from "../components/Grid";
import Icon, { type IconName } from "../components/Icon";
import ProductCard from "../components/ProductCard";
import ProductImage from "../components/ProductImage";
import TintCard, { TINTS } from "../components/TintCard";
import Weights from "../components/Weights";

const FEATURE_ICONS: IconName[] = ["card", "returns", "chat"];

/**
 * THE HOME PAGE. No photograph opens it: a heading over three lines whose
 * words change weight, with the dark pill sitting ON its last line and a
 * short block of proof at its right — and under that, three tall tinted
 * cards. The colour on this page is in the cards and nowhere else.
 *
 * Then the departments as grey wordmarks between two hairlines, the two
 * things the shop leads with, a quiet grey band listing where to start, a
 * peach row of three reassurances, and everything else as small cards, four
 * to a row. A band with nothing in it is not drawn.
 */
export default function Home() {
  usePageMeta("pages.home.title", "pages.home.description");
  const lead = useCollection<Product>("products", { where: { status: "live", deal: true }, pageSize: 5 }).items;
  const picks = useCollection<Product>("products", { where: { status: "live" }, pageSize: 8 });
  const departments = useCollection<Department>("categories", { pageSize: 12 }).items;
  const features = useList<Feature>("home.features");
  const nameOf = (id: string) => departments.find((d) => d.id === id)?.name;
  const month = lead.slice(3, 5);
  return (
    <>
      <section className="flex flex-col gap-10 pb-12 pt-14 lg:flex-row lg:items-center lg:justify-between lg:pt-20">
        <h1 className="text-[length:var(--size-h1)] leading-[1.18] text-ink">
          <T k="hero.headline.light" as="span" className="block font-light" />
          <T k="hero.headline.bold" as="span" className="block font-bold" />
          {/* THE PILL SITS ON THE LAST LINE, beside its words — not under
              the heading, where every other page puts it. */}
          <span className="flex flex-wrap items-center gap-x-8 gap-y-4">
            <T k="hero.headline.tail" className="font-light" />
            <Link to="/shop" data-testid="hero-cta" data-magnetic=""
              className="inline-flex rounded-full bg-accent px-10 py-3.5 font-body text-sm font-medium tracking-normal text-accent-ink hover:opacity-90">
              <T k="hero.cta.label" />
            </Link>
          </span>
        </h1>
        {/* THE PROOF: three of the shop's own things as overlapping rounds,
            how many there are — counted, not claimed — and two lines. */}
        {picks.total > 0 && (
          <div className="flex items-start gap-5 lg:w-[23rem]">
            <span className="mt-12 hidden h-px w-16 shrink-0 bg-ink lg:block" />
            <div>
              <div className="flex items-center gap-3">
                <span className="flex -space-x-3">
                  {picks.items.slice(0, 3).map((p) => (
                    <span key={p.id} className="block w-10 overflow-hidden rounded-full ring-2 ring-[var(--ground)]">
                      <ProductImage src={p.image} alt="" aspect="1/1" className="rounded-none" />
                    </span>
                  ))}
                </span>
                <span className="text-sm font-bold leading-tight text-ink">
                  <span className="tabular-nums" data-count="">{picks.total}</span>{" "}<T k="hero.proof.count" />
                </span>
              </div>
              <T k="hero.proof.line" as="p" className="mt-3 max-w-[22ch] text-sm leading-snug text-ink" />
            </div>
          </div>
        )}
      </section>

      {lead.length > 0 && (
        <Grid max={3} phone={1}>
          {lead.slice(0, 3).map((p, i) => <TintCard key={p.id} product={p} tint={i} label={nameOf(p.category)} />)}
        </Grid>
      )}

      {/* The departments as wordmarks, between two hairlines. */}
      {departments.length > 0 && (
        <nav aria-label="Departments" className="mt-16 flex flex-wrap items-center justify-between gap-x-10 gap-y-4 border-y border-line py-9">
          {departments.slice(0, 5).map((d) => (
            <Link key={d.id} to={`/shop?category=${d.id}`} data-record={`categories:${d.id}`} data-testid={`mark-${d.id}`}
              className="text-lg font-semibold uppercase tracking-[0.18em] text-ink-soft hover:text-ink">{d.name}</Link>
          ))}
        </nav>
      )}

      {month.length > 0 && (
        <section className="flex flex-col gap-10 pt-20 lg:flex-row lg:items-center lg:gap-16">
          <div className="flex items-end gap-5 lg:w-7/12">
            {month.map((p, i) => (
              <div key={p.id} className={i === 0 ? "w-5/12" : "w-7/12 lg:w-6/12"}>
                <TintCard product={p} tint={i === 0 ? 3 : 0} compact />
              </div>
            ))}
          </div>
          <div className="lg:w-5/12">
            <h2 className="text-[length:var(--size-h1)] leading-[1.12] text-ink"><Weights k="home.month" /></h2>
            <T k="home.month.line" as="p" className="mt-5 max-w-[34ch] text-sm leading-relaxed text-ink-soft" />
          </div>
        </section>
      )}

      {departments.length > 0 && <Start departments={departments.slice(0, 3)} />}

      {features.length > 0 && (
        <section className={`mt-16 flex flex-col gap-8 rounded px-6 py-10 md:flex-row ${TINTS[3]}`}>
          {features.slice(0, 3).map((f, i) => (
            <div key={f.bold} className="flex flex-1 flex-col items-center text-center">
              <span className="grid h-12 w-12 place-items-center rounded-full bg-ground text-ink"><Icon name={FEATURE_ICONS[i]} size={20} /></span>
              <T k={`home.features.${i}.bold`} as="p" className="mt-3 text-sm font-semibold text-ink" />
              <T k={`home.features.${i}.light`} as="p" className="mt-1 max-w-[24ch] text-xs leading-relaxed text-ink-soft" />
            </div>
          ))}
        </section>
      )}

      {picks.items.length > 0 && (
        <section className="pt-20">
          <div className="mb-8 text-center">
            <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.picks" /></h2>
            <T k="home.picks.line" as="p" className="mx-auto mt-3 max-w-[32ch] text-sm leading-relaxed text-ink-soft" />
            <nav aria-label="Departments" className="mt-6 flex flex-wrap justify-center gap-2 text-xs">
              <Link to="/shop" data-testid="picks-all" className="rounded-full bg-accent px-4 py-1.5 font-medium text-accent-ink"><T k="home.picks.all" /></Link>
              {departments.slice(0, 4).map((d) => (
                <Link key={d.id} to={`/shop?category=${d.id}`} data-testid={`chip-${d.id}`}
                  className="rounded-full border border-ink px-4 py-1.5 text-ink hover:bg-accent hover:text-accent-ink">{d.name}</Link>
              ))}
            </nav>
          </div>
          <Grid max={4}>{picks.items.map((p) => <ProductCard key={p.id} product={p} />)}</Grid>
        </section>
      )}
    </>
  );
}

/**
 * Where to start: a quiet grey band — a heading and three rows at the left,
 * each a department with a line about it and an arrow, a hairline between;
 * and a photograph at the right when there is one. With none the rows take
 * the band's width.
 */
function Start({ departments }: { departments: Department[] }) {
  const image = useContent("home.start.image");
  return (
    <section className="mt-20 flex flex-col gap-10 rounded bg-surface p-6 lg:flex-row lg:items-stretch lg:gap-14 lg:p-12">
      <div className="flex-1">
        <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="home.start" /></h2>
        <T k="home.start.line" as="p" className="mt-3 max-w-[40ch] text-sm leading-relaxed text-ink-soft" />
        <ul className="mt-8">
          {departments.map((d) => (
            <li key={d.id} className="border-b border-line last:border-b-0">
              <Link to={`/shop?category=${d.id}`} data-testid={`start-${d.id}`} className="group flex items-center justify-between gap-6 py-5">
                <span>
                  <span className="block font-medium text-ink underline-offset-4 group-hover:underline">{d.name}</span>
                  <T k="home.start.row" as="span" className="mt-1 block text-xs text-ink-soft" />
                </span>
                <Icon name="arrow" size={18} className="shrink-0 text-ink" />
              </Link>
            </li>
          ))}
        </ul>
      </div>
      {image && (
        <div className="overflow-hidden rounded lg:w-5/12">
          <Picture src={image} alt="" width={900} height={1100} className="aspect-[4/5] h-full w-full object-cover" />
        </div>
      )}
    </section>
  );
}
