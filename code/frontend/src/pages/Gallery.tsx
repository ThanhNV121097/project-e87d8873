import { Link } from "react-router-dom";
import { T } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { Picture } from "../picture";
import type { Department, Shot } from "../lib/shop";
import Icon from "../components/Icon";
import Weights from "../components/Weights";

/**
 * WHERE THESE BIKES GO: the riding a buyer is choosing a bike for, shown
 * large. The heading at the left with the first photograph wide beside it;
 * then the rest two to a row, every frame the same landscape shape, each
 * captioned under a hairline with the riding type it belongs to — a link to
 * the bikes for it.
 */
export default function Gallery() {
  usePageMeta("pages.gallery.title", "pages.gallery.description");
  const { items, total } = useCollection<Shot>("gallery", { pageSize: 24 });
  const types = useCollection<Department>("categories", { pageSize: 12 }).items;
  const [first, ...rest] = items;
  const typeOf = (id?: string) => types.find((d) => d.id === id);
  return (
    <div className="pb-8 pt-10">
      <section className="grid gap-10 lg:grid-cols-12 lg:items-end">
        <div className="lg:col-span-4">
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.gallery.heading" /></h1>
          <T k="pages.gallery.lead" as="p" className="mt-5 max-w-[40ch] text-sm leading-relaxed text-ink-soft" />
          <p className="mt-6 text-sm text-ink-soft"><span className="text-2xl font-bold tabular-nums text-ink">{total}</span> <T k="pages.gallery.count" /></p>
        </div>
        {first && (
          <figure data-record={`gallery:${first.id}`} className="lg:col-span-8">
            <div className="overflow-hidden rounded">
              <Picture src={first.image} alt={first.title} width={1280} height={853} loading="eager" motion="parallax" className="aspect-[3/2] w-full object-cover" />
            </div>
            <Caption shot={first} type={typeOf(first.type)} />
          </figure>
        )}
      </section>

      <section className="grid gap-x-8 gap-y-14 pt-24 md:grid-cols-2">
        {rest.map((s, i) => {
          const wide = rest.length % 2 === 1 && i === rest.length - 1;
          return (
            <figure key={s.id} data-record={`gallery:${s.id}`} className={wide ? "md:col-span-2" : ""}>
              <div className="overflow-hidden rounded">
                <Picture src={s.image} alt={s.title} width={1280} height={853} className={`w-full object-cover ${wide ? "aspect-[3/2] md:aspect-[21/9]" : "aspect-[3/2]"}`} />
              </div>
              <Caption shot={s} type={typeOf(s.type)} />
            </figure>
          );
        })}
      </section>

      <section className="mt-24 flex flex-col gap-6 border-y border-line py-10 lg:flex-row lg:items-center lg:justify-between">
        <p className="text-[length:var(--size-h2)] leading-tight text-ink"><Weights k="pages.gallery.close" /></p>
        <Link to="/shop" data-testid="gallery-cta" className="inline-flex items-center gap-2 self-start whitespace-nowrap rounded-full bg-accent px-8 py-3.5 text-sm font-medium text-accent-ink lg:self-auto">
          <T k="pages.gallery.cta" /><Icon name="arrow" size={16} />
        </Link>
      </section>
    </div>
  );
}

function Caption({ shot, type }: { shot: Shot; type?: Department }) {
  return (
    <figcaption className="mt-4 flex items-start justify-between gap-6 border-t border-line pt-4">
      <span>
        <span className="block font-semibold text-ink">{shot.title}</span>
        {shot.caption && <span className="mt-1 block max-w-[52ch] text-sm text-ink-soft">{shot.caption}</span>}
      </span>
      {type && (
        <Link to={`/shop?category=${type.id}`} className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap rounded-full border border-ink px-4 py-1.5 text-xs text-ink hover:bg-accent hover:text-accent-ink">
          {type.name}<Icon name="next" size={13} />
        </Link>
      )}
    </figcaption>
  );
}
