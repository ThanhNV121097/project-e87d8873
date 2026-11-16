import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { useCollection } from "../data";
import Icon from "./Icon";
import ProductImage from "./ProductImage";
import SaveButton from "./SaveButton";
import StockLine from "./StockLine";
import { formatPrice } from "../lib/format";
import type { Department, Product } from "../lib/shop";

/** Which of a bike's own photographs a card leads with. */
export type CardView = "drive" | "whole" | "front";

const ROTATION: readonly CardView[] = ["drive", "whole", "front"];

function pictureFor(p: Product, view: CardView) {
  if (view === "drive") return p.image_back;
  if (view === "front") return p.image_detail;
  return p.image;
}

/**
 * THE VIEW EACH CARD IN A ROW LEADS WITH. Every mountain bike on its maker's
 * white sweep is the same silhouette — two wheels and a triangle — so a row
 * of them is told apart only by reading. The maker shot every bike three
 * ways, so a row takes turns: the drivetrain close-up, the whole bike, the
 * fork and cockpit — the back end, the bike, the front end — and each card
 * shows the whole bike under the pointer. A bike with only its whole
 * photograph shows that, and the turn carries on after it, so two whole
 * bikes never stand side by side.
 */
export function cardViews(list: readonly Product[]): CardView[] {
  let at = 0;
  return list.map((p) => {
    let view = ROTATION[at % ROTATION.length];
    if (!pictureFor(p, view)) view = "whole";
    at = ROTATION.indexOf(view) + 1;
    return view;
  });
}

/**
 * A BIKE IN A GRID: one of the maker's photographs on a white 4:3 tile with a
 * hairline edge, a heart at its top right; then the brand and riding type in
 * small capitals, the name with the one line a rider compares by under it at
 * the left, and at the right a column of its own for the price with the old
 * price struck through beneath it — both right-aligned, however long the line
 * beside them runs; the stock in the owner's words, and an outlined pill to
 * the sizes.
 */
export default function ProductCard({ product, view = "whole" }: { product: Product; view?: CardView }) {
  const types = useCollection<Department>("categories", { pageSize: 12 }).items;
  const type = types.find((t) => t.id === product.category)?.name;
  const out = (product.stock ?? 0) <= 0;
  const lead = pictureFor(product, view) || product.image;
  const second = lead === product.image ? product.image_back : product.image;
  const viewWord = useContent(`card.view.${view}`);
  const name = `${product.brand} ${product.name}`;
  return (
    <article data-record={`products:${product.id}`} className="relative flex h-full flex-col">
      <Link to={`/shop/${product.id}`} data-testid={`card-${product.id}`} className="group block">
        <ProductImage src={lead} hover={second} alt={view === "whole" || !viewWord ? name : `${name}, ${viewWord}`} aspect="4/3"
          className="rounded border border-line" />
        <span className="mt-4 block text-xs uppercase tracking-[0.14em] text-ink-soft">{product.brand}{type ? ` · ${type}` : ""}</span>
        <span className="mt-1.5 flex items-start justify-between gap-4">
          <span className="min-w-0">
            <span className="block text-base font-semibold text-ink underline-offset-4 group-hover:underline">{product.name}</span>
            {product.fact && <span className="mt-1 block text-sm text-ink-soft">{product.fact}</span>}
          </span>
          <span className="flex shrink-0 flex-col items-end text-right">
            <span className="text-base font-bold tabular-nums text-ink">{formatPrice(product.price)}</span>
            {product.was !== undefined && product.was > product.price && (
              <span data-testid={`card-was-${product.id}`} className="mt-1 text-sm tabular-nums text-ink-soft line-through">{formatPrice(product.was)}</span>
            )}
          </span>
        </span>
      </Link>
      <StockLine stock={product.stock ?? 0} className="mt-3" />
      <div className="mt-4 flex flex-1 items-end">
        <Link to={`/shop/${product.id}#sizes`}
          className="inline-flex w-full items-center justify-center gap-2 rounded-full border border-ink py-2.5 text-sm font-medium text-ink hover:bg-accent hover:text-accent-ink">
          <T k={out ? "card.ask" : "card.choose"} /><Icon name="next" size={15} />
        </Link>
      </div>
      <SaveButton id={product.id} name={product.name} className="absolute right-3 top-3" />
    </article>
  );
}
