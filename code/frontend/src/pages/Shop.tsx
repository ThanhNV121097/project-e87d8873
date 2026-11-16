import { useState } from "react";
import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import { useTableQuery } from "../lib/table";
import { inBand, PRICE_BANDS, type Department, type Product } from "../lib/shop";
import FilterRail from "../components/FilterRail";
import Grid from "../components/Grid";
import Icon from "../components/Icon";
import Bleed from "../components/Bleed";
import ProductCard, { cardViews } from "../components/ProductCard";
import Select from "../components/Select";
import Weights from "../components/Weights";

const SIZES = ["XS", "S", "M", "L", "XL"];
const PAGE = 9;

/**
 * ALL THE BIKES. The heading changes weight like every head on the site and
 * the count stands at its right; under it the riding types as pills (the
 * template's "Favorite Picks" chips), the chosen one filled dark. Down the
 * left, what a rider narrows a bike by: frame size, wheel, brand, price.
 * Fourteen bikes are few, so they are shown large — three to a row, each
 * with the figures to choose from without opening it.
 *
 * Everything that narrows the list lives in the address (useTableQuery), so
 * "trail bikes in M under 70 million" is a link somebody can send, and coming
 * back from a bike lands on the same list.
 */
export default function Shop() {
  usePageMeta("pages.shop.title", "pages.shop.description");
  const t = useTableQuery({ size: 100 });
  const [shown, setShown] = useState(PAGE);
  const types = useCollection<Department>("categories", { pageSize: 24 }).items;
  const chosen = t.list("category");
  const sizes = t.list("size");
  const bands = t.list("price");
  const { items, loading } = useCollection<Product>("products", {
    where: { status: "live", category: chosen, brand: t.list("brand"), wheel: t.list("wheel") },
    q: t.q, search: ["name", "brand", "category", "fact", "description", "frame", "fork", "drivetrain", "travel"],
    sort: t.sort, pageSize: 100,
  });
  const all = useCollection<Product>("products", { where: { status: "live" }, pageSize: 100 }).items;
  const found = items.filter((p) =>
    (sizes.length === 0 || sizes.some((s) => (p.sizes ?? []).includes(s))) &&
    (bands.length === 0 || bands.some((b) => inBand(p.price, b))));
  const views = cardViews(found);
  const one = chosen.length === 1 ? types.find((d) => d.id === chosen[0]) : undefined;
  const brands = [...new Set(all.map((p) => p.brand))].sort();
  const sortLabels = [useContent("pages.shop.sort_default"), useContent("pages.shop.sort_low"), useContent("pages.shop.sort_high")];
  const wheel29 = useContent("pages.shop.wheel_29");
  const wheelMixed = useContent("pages.shop.wheel_mixed");
  const bandWords = useContent<Record<string, string>>("pages.shop.band");
  const bandLabels = PRICE_BANDS.map((b) => bandWords?.[b.id] ?? b.id);
  const sortLabel = useContent("pages.shop.sort");
  const image = useContent("pages.shop.image");
  const imageAlt = useContent("pages.shop.image_alt");
  const pick = (id: string) => { t.set("category", chosen.length === 1 && chosen[0] === id ? [] : [id]); setShown(PAGE); };
  return (
    <div className="pb-8">
      {/* The head of the catalogue on a band that runs to both screen edges:
          where you are and the riding types to choose from at the left, and
          the country these bikes are for off the right edge (on a phone, a
          strip across the top of the band). */}
      <Bleed className="bg-surface lg:flex lg:flex-row-reverse lg:gap-12">
      {image && (
        <div className="relative -mx-[var(--gutter)] aspect-[21/9] overflow-hidden lg:mx-0 lg:w-4/12 lg:shrink-0 lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:aspect-auto">
          <Picture src={image} alt={imageAlt} width={1280} height={730} loading="eager" className="absolute inset-0 h-full w-full object-cover" />
        </div>
      )}
      <div className="min-w-0 pb-8 pt-8 lg:flex-1">
      <nav aria-label="Breadcrumb" className="mb-6 flex items-center gap-1.5 text-sm text-ink-soft">
        <Link to="/" className="hover:text-ink"><Icon name="home" size={16} label="Home" /></Link>
        <Icon name="next" size={14} />
        <Link to="/shop" className="hover:text-ink"><T k="pages.shop.crumb" /></Link>
        {one && (<><Icon name="next" size={14} /><span className="text-ink">{one.name}</span></>)}
      </nav>

      <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink">
            {one ? <><span className="font-light">{one.name} </span><span className="font-bold"><T k="pages.shop.bikes_word" /></span></> : <Weights k="pages.shop.heading" />}
          </h1>
          <p className="mt-5 max-w-[56ch] text-sm leading-relaxed text-ink-soft">{one ? one.line : <T k="pages.shop.lead" />}</p>
        </div>
      </div>

      <nav aria-label="Riding types" className="mt-8 flex flex-wrap gap-2 text-sm">
        <button type="button" onClick={() => { t.set("category", []); setShown(PAGE); }} data-testid="chip-all"
          className={`rounded-full px-5 py-2 ${chosen.length === 0 ? "bg-accent font-medium text-accent-ink" : "border border-ink text-ink hover:bg-wash"}`}>
          <T k="pages.shop.all" />
        </button>
        {types.map((d) => {
          const on = chosen.includes(d.id);
          return (
            <button key={d.id} type="button" onClick={() => pick(d.id)} aria-pressed={on} data-testid={`chip-${d.id}`}
              className={`rounded-full px-5 py-2 ${on ? "bg-accent font-medium text-accent-ink" : "border border-ink text-ink hover:bg-wash"}`}>
              {d.name}
            </button>
          );
        })}
      </nav>
      </div>
      </Bleed>

      {t.q && (
        <p className="mt-6 flex items-center gap-3 text-sm text-ink-soft">
          <Icon name="search" size={16} /><T k="pages.shop.searching" />{" "}<span className="font-medium text-ink">“{t.q}”</span>
          <button type="button" onClick={() => t.setQ("")} className="underline underline-offset-4 hover:text-ink"><T k="pages.shop.clear_search" /></button>
        </p>
      )}

      <div className="mt-10 flex flex-col gap-10 lg:flex-row lg:gap-12">
        <FilterRail className="lg:w-52 lg:shrink-0" clear={<T k="pages.shop.clear" />}
          groups={[
            { key: "size", label: <T k="pages.shop.filter_size" />, options: SIZES.map((s) => [s, s] as const) },
            { key: "wheel", label: <T k="pages.shop.filter_wheel" />, options: [["29", wheel29], ["mixed", wheelMixed]] as const },
            { key: "brand", label: <T k="pages.shop.filter_brand" />, options: brands.map((b) => [b, b] as const) },
            { key: "price", label: <T k="pages.shop.filter_price" />, options: PRICE_BANDS.map((b, i) => [b.id, bandLabels[i]] as const) },
          ]} />
        <div className="min-w-0 flex-1">
          {/* How many, and in what order — over the grid they describe. */}
          <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
            <p data-testid="shop-count" className="text-sm text-ink-soft">
              <span className="text-2xl font-bold tabular-nums text-ink">{found.length}</span>{" "}<T k="pages.shop.count" />
            </p>
            <Select aria-label={sortLabel} value={t.sort} onChange={(v) => t.set("sort", v)}
              options={[{ value: "", label: sortLabels[0] }, { value: "price", label: sortLabels[1] }, { value: "-price", label: sortLabels[2] }]}
              className="w-52" />
          </div>
          {loading ? (
            <p className="py-16 text-sm text-ink-soft"><T k="pages.shop.loading" /></p>
          ) : found.length === 0 ? (
            <div className="flex flex-col items-start gap-5 rounded bg-[var(--tint-1)] px-8 py-14">
              <T k="pages.shop.empty" as="p" className="text-xl text-ink" />
              <T k="pages.shop.empty_line" as="p" className="max-w-[50ch] text-sm text-ink-soft" />
              <button type="button" onClick={t.clear} data-testid="shop-clear" className="rounded-full bg-accent px-6 py-2.5 text-sm font-medium text-accent-ink">
                <T k="pages.shop.clear" />
              </button>
            </div>
          ) : (
            <>
              <Grid max={3} phone={1}>{found.slice(0, shown).map((p, i) => <ProductCard key={p.id} product={p} view={views[i]} />)}</Grid>
              {found.length > shown && (
                <div className="mt-12 flex justify-center">
                  <button type="button" onClick={() => setShown((n) => n + PAGE)} data-testid="shop-more"
                    className="rounded-full border border-ink px-8 py-3 text-sm font-medium text-ink hover:bg-accent hover:text-accent-ink">
                    <T k="pages.shop.more" /> <span className="tabular-nums text-ink-soft">({found.length - shown})</span>
                  </button>
                </div>
              )}
            </>
          )}
          <aside className="-mx-[var(--gutter)] mt-16 flex flex-col gap-5 bg-[var(--tint-3)] px-6 py-8 sm:mx-0 sm:rounded lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:flex-row lg:items-center lg:justify-between lg:rounded-r-none lg:pl-10 lg:pr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter)+2.5rem)]">
            <div>
              <p className="text-xl leading-tight text-ink"><Weights k="pages.shop.help" /></p>
              <T k="pages.shop.help.line" as="p" className="mt-2 max-w-[52ch] text-sm text-ink-soft" />
            </div>
            <Link to="/contact" className="inline-flex shrink-0 items-center gap-2 self-start whitespace-nowrap rounded-full bg-accent px-6 py-3 text-sm font-medium text-accent-ink lg:self-auto">
              <T k="pages.shop.help.cta" /><Icon name="arrow" size={16} />
            </Link>
          </aside>
        </div>
      </div>
    </div>
  );
}
