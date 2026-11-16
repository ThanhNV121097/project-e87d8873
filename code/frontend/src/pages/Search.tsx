import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { T, useContent, useList } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import { formatPrice } from "../lib/format";
import { useTableQuery } from "../lib/table";
import type { Department, Faq, Product, Service, Shot } from "../lib/shop";
import Grid from "../components/Grid";
import Icon from "../components/Icon";
import ProductCard, { cardViews } from "../components/ProductCard";
import { TINTS } from "../components/TintCard";
import Weights from "../components/Weights";

const KINDS = ["bikes", "types", "services", "faqs", "gallery"] as const;
type Kind = (typeof KINDS)[number];
const BIKES_SHOWN = 6;

/**
 * SEARCH THE WHOLE SHOP: what the magnifier in the header opens. One field,
 * set as the template's pill with the dark button on its line, and under it
 * the kinds of thing found as chips with their counts — Bikes, Riding types,
 * Workshop, Questions, Gallery — the chosen one filled dark. The words and
 * the chosen kind live in the address (`/search?q=bleed&in=services`), so a
 * search is a link somebody can send.
 *
 * Each kind answers on this page: a bike as its card with its price, a
 * workshop job with its time and price at the right, a question with its
 * answer, a photograph with its caption — and a link to where it lives.
 */
export default function Search() {
  usePageMeta("pages.search.title", "pages.search.description");
  const t = useTableQuery();
  const q = t.q.trim();
  const on = q !== "";
  const [text, setText] = useState(q);
  useEffect(() => { setText(q); }, [q]);
  const label = useContent("header.search");
  const placeholder = useContent("pages.search.placeholder");
  const filterLabel = useContent("pages.search.filter");
  const image = useContent("pages.search.image");
  const imageAlt = useContent("pages.search.image_alt");

  const bikes = useCollection<Product>("products", {
    where: { status: "live" }, q, pageSize: BIKES_SHOWN,
    search: ["name", "brand", "category", "fact", "description", "frame", "fork", "drivetrain", "travel"],
  });
  const types = useCollection<Department>("categories", { q, search: ["name", "line", "terrain"], pageSize: 12 });
  const services = useCollection<Service>("services", { q, search: ["name", "line"], pageSize: 24 });
  const faqs = useCollection<Faq>("faqs", { q, search: ["q", "a"], pageSize: 24 });
  const shots = useCollection<Shot>("gallery", { q, search: ["title", "caption", "type"], pageSize: 12 });
  const counts: Record<Kind, number> = {
    bikes: bikes.total, types: types.total, services: services.total, faqs: faqs.total, gallery: shots.total,
  };
  const sum = KINDS.reduce((n, k) => n + counts[k], 0);
  const bikeViews = cardViews(bikes.items);
  const asked = t.get("in") as Kind | "";
  // A kind chosen for an earlier search that finds nothing for this one shows everything.
  const kind = asked && KINDS.includes(asked) && counts[asked] > 0 ? asked : "";
  const show = (k: Kind) => on && counts[k] > 0 && (kind === "" || kind === k);
  const loading = bikes.loading || services.loading || faqs.loading;

  const submit = (e: FormEvent) => {
    e.preventDefault();
    t.setQ(text);
  };

  return (
    <div className="pb-8 pt-10">
      {/* The head: heading, field and the kinds found down the left, and the
          page's own riding photograph beside them, run off the right edge —
          in every state, typed or not (on a phone, edge to edge under it). */}
      <section className="grid gap-8 lg:grid-cols-12 lg:gap-12">
        <div className="lg:col-span-7">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.search.heading" /></h1>
          <T k="pages.search.lead" as="p" className="mt-5 max-w-[56ch] text-sm leading-relaxed text-ink-soft" />
        </div>
        {on && !loading && (
          <p data-testid="search-count" className="shrink-0 text-sm text-ink-soft">
            <span className="text-2xl font-bold tabular-nums text-ink">{sum}</span>{" "}<T k="pages.search.count" />
          </p>
        )}
        </div>

      {/* The field is the pill: its own edge, the dark button set inside its line. */}
      <form role="search" onSubmit={submit} className="relative mt-8">
        <span className="pointer-events-none absolute left-5 top-1/2 -translate-y-1/2 text-ink-soft"><Icon name="search" size={20} /></span>
        <input id="site-search" data-testid="search-page-input" value={text} onChange={(e) => setText(e.target.value)}
          aria-label={label} placeholder={placeholder}
          className="w-full appearance-none text-ellipsis rounded-full border border-ink bg-ground py-4 pl-12 pr-36 text-base text-ink placeholder:text-ink-soft focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-ink lg:text-lg" />
        <button type="submit" data-testid="search-go"
          className="absolute right-1.5 top-1/2 inline-flex -translate-y-1/2 items-center gap-2 whitespace-nowrap rounded-full bg-accent px-6 py-3 text-sm font-medium text-accent-ink hover:opacity-90">
          <T k="pages.search.go" /><Icon name="arrow" size={16} />
        </button>
      </form>

      {on && sum > 0 && (
        <nav aria-label={filterLabel} className="mt-6 flex flex-wrap gap-2 text-sm">
          <Chip on={kind === ""} onPick={() => t.set("in", "")} testid="search-in-all">
            <T k="pages.search.all" /> <span className="tabular-nums opacity-70">{sum}</span>
          </Chip>
          {KINDS.filter((k) => counts[k] > 0).map((k) => (
            <Chip key={k} on={kind === k} onPick={() => t.set("in", k)} testid={`search-in-${k}`}>
              <T k={`pages.search.kinds.${k}`} /> <span className="tabular-nums opacity-70">{counts[k]}</span>
            </Chip>
          ))}
        </nav>
      )}
        </div>
        {image && (
          <div className="relative -mx-[var(--gutter)] aspect-[16/9] overflow-hidden lg:col-span-5 lg:mx-0 lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:aspect-auto lg:min-h-[16rem] lg:rounded-l">
            <Picture src={image} alt={imageAlt} width={1280} height={960} loading="eager" className="absolute inset-0 h-full w-full object-cover" />
          </div>
        )}
      </section>

      {/* Nothing typed yet: what people search for, and the four riding
          types to start from instead. */}
      {!on && <StartHere types={types.items} />}

      {on && !loading && sum === 0 && <Nothing q={q} />}

      {show("bikes") && (
        <Block kind="bikes" count={bikes.total} more={`/shop?q=${encodeURIComponent(q)}`}>
          <Grid max={3} phone={1}>{bikes.items.map((p, i) => <ProductCard key={p.id} product={p} view={bikeViews[i]} />)}</Grid>
        </Block>
      )}

      {show("types") && (
        <Block kind="types" count={types.total} more="/shop">
          <ul className="rounded bg-surface px-6 lg:px-10">
            {types.items.map((d) => (
              <li key={d.id} data-record={`categories:${d.id}`} className="border-b border-line last:border-b-0">
                <Link to={`/shop?category=${d.id}`} className="group flex items-center justify-between gap-6 py-5">
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
        </Block>
      )}

      {show("services") && (
        <Block kind="services" count={services.total} more="/about#workshop">
          <table data-testid="search-services" className="w-full text-sm">
            <thead>
              <tr className="border-b border-ink text-left text-xs uppercase tracking-wider text-ink-soft">
                <th className="pb-3 font-medium"><T k="pages.about.col_job" /></th>
                <th className="hidden pb-3 text-right font-medium sm:table-cell"><T k="pages.about.col_time" /></th>
                <th className="pb-3 text-right font-medium"><T k="pages.about.col_price" /></th>
              </tr>
            </thead>
            <tbody>
              {services.items.map((s) => (
                <tr key={s.id} data-record={`services:${s.id}`} className="border-b border-line align-top">
                  <td className="py-4 pr-4">
                    <Link to={`/about#job-${s.id}`} className="block font-semibold text-ink underline-offset-4 hover:underline">{s.name}</Link>
                    <span className="mt-1 block text-xs text-ink-soft">{s.line}</span>
                  </td>
                  <td className="hidden py-4 text-right text-ink-soft sm:table-cell">{s.time}</td>
                  <td className="py-4 text-right font-semibold tabular-nums text-ink">{s.price > 0 ? formatPrice(s.price) : <T k="pages.about.included" />}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Block>
      )}

      {show("faqs") && (
        <Block kind="faqs" count={faqs.total} more="/about#faq">
          <div data-testid="search-faqs" className="divide-y divide-line border-y border-line">
            {faqs.items.map((f) => (
              <article key={f.id} data-record={`faqs:${f.id}`} className="flex flex-col gap-2 py-5 lg:flex-row lg:gap-12">
                <Link to={`/about#faq-${f.id}`} className="font-semibold text-ink underline-offset-4 hover:underline lg:w-5/12 lg:shrink-0">{f.q}</Link>
                <p className="max-w-[65ch] text-sm leading-relaxed text-ink-soft">{f.a}</p>
              </article>
            ))}
          </div>
        </Block>
      )}

      {show("gallery") && (
        <Block kind="gallery" count={shots.total} more="/gallery">
          <Grid max={3} phone={1}>
            {shots.items.map((s) => (
              <Link key={s.id} to={`/gallery#shot-${s.id}`} data-record={`gallery:${s.id}`} className="group block">
                <span className="block overflow-hidden rounded">
                  <Picture src={s.image} alt={s.title} width={1280} height={853} className="aspect-[3/2] w-full object-cover" />
                </span>
                <span className="mt-3 block font-semibold text-ink underline-offset-4 group-hover:underline">{s.title}</span>
                {s.caption && <span className="mt-1 block text-sm text-ink-soft">{s.caption}</span>}
              </Link>
            ))}
          </Grid>
        </Block>
      )}
    </div>
  );
}

/**
 * Nothing typed yet: what people search for on a tinted panel, and the four
 * riding types to start from.
 */
function StartHere({ types }: { types: Department[] }) {
  const suggestions = useList<string>("pages.search.suggestions");
  return (
    <>
        <div className="mt-12 flex flex-col items-start justify-center gap-5 rounded bg-[var(--tint-1)] px-6 py-10 lg:px-10">
          <T k="pages.search.prompt" as="p" className="text-[length:var(--size-h2)] font-light leading-tight text-ink" />
          <T k="pages.search.prompt_line" as="p" className="max-w-[52ch] text-sm text-ink-soft" />
          <div className="flex flex-wrap gap-2 text-sm">
            {suggestions.map((s, i) => (
              <Link key={s} to={`/search?q=${encodeURIComponent(s)}`} data-testid={`search-suggest-${i}`}
                className="rounded-full bg-ground px-4 py-2 text-ink hover:bg-accent hover:text-accent-ink">{s}</Link>
            ))}
          </div>
        </div>
      {types.length > 0 && (
        <section className="pt-16">
          <h2 className="mb-8 text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="pages.search.start" /></h2>
          <Grid max={4} phone={1}>
            {types.map((d, i) => (
              <Link key={d.id} to={`/shop?category=${d.id}`} data-record={`categories:${d.id}`} data-testid={`search-type-${d.id}`}
                className={`group flex h-full flex-col gap-3 rounded p-6 ${TINTS[i % TINTS.length]}`}>
                <span className="text-[1.75rem] font-bold leading-none tracking-tight text-ink">{d.name}</span>
                <span className="flex-1 text-sm leading-relaxed text-ink">{d.line}</span>
                <span className="flex items-center justify-between gap-4 border-t border-line pt-3 text-xs text-ink-soft">
                  {d.terrain}
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-accent text-accent-ink"><Icon name="arrow" size={15} /></span>
                </span>
              </Link>
            ))}
          </Grid>
        </section>
      )}
    </>
  );
}

/** Nothing found: say what was looked for, and the two ways on. */
function Nothing({ q }: { q: string }) {
  return (
    <div data-testid="search-none" className="mt-12 flex flex-col items-start gap-5 rounded bg-[var(--tint-1)] px-6 py-10 lg:px-10">
      <p className="text-xl text-ink"><T k="pages.search.none" />{" "}<span className="font-semibold">“{q}”</span></p>
      <T k="pages.search.none_line" as="p" className="max-w-[52ch] text-sm text-ink-soft" />
      <div className="flex flex-wrap gap-3">
        <Link to="/contact" className="inline-flex items-center gap-2 whitespace-nowrap rounded-full bg-accent px-6 py-3 text-sm font-medium text-accent-ink">
          <T k="pages.search.ask" /><Icon name="arrow" size={16} />
        </Link>
        <Link to="/shop" className="inline-flex items-center whitespace-nowrap rounded-full border border-ink px-6 py-3 text-sm font-medium text-ink hover:bg-ground">
          <T k="pages.search.browse" />
        </Link>
      </div>
    </div>
  );
}

/** One kind of result: its name in bold and its count light, the way to where it lives at the right. */
function Block({ kind, count, more, children }: { kind: Kind; count: number; more: string; children: ReactNode }) {
  return (
    <section data-testid={`search-block-${kind}`} className="pt-16">
      <div className="mb-6 flex items-end justify-between gap-6 border-b border-line pb-4">
        <h2 className="text-[length:var(--size-h2)] leading-tight text-ink">
          <T k={`pages.search.kinds.${kind}`} className="font-bold" />{" "}<span className="font-light tabular-nums">{count}</span>
        </h2>
        <Link to={more} className="inline-flex items-center gap-1 whitespace-nowrap text-sm font-medium text-ink underline-offset-4 hover:underline">
          <T k={`pages.search.more.${kind}`} /><Icon name="next" size={15} />
        </Link>
      </div>
      {children}
    </section>
  );
}

function Chip({ on, onPick, testid, children }: { on: boolean; onPick: () => void; testid: string; children: ReactNode }) {
  return (
    <button type="button" onClick={onPick} aria-pressed={on} data-testid={testid}
      className={`whitespace-nowrap rounded-full px-5 py-2 ${on ? "bg-accent font-medium text-accent-ink" : "border border-ink text-ink hover:bg-wash"}`}>
      {children}
    </button>
  );
}
