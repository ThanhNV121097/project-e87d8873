import { Link } from "react-router-dom";
import { T } from "../editable";
import { useCollection } from "../data";
import Icon from "./Icon";
import ProductImage from "./ProductImage";
import SaveButton from "./SaveButton";
import StockLine from "./StockLine";
import { formatPrice } from "../lib/format";
import type { Department, Product } from "../lib/shop";

/**
 * A BIKE IN A GRID: the maker's photograph on a white 4:3 tile with a
 * hairline edge (the tile and the photograph's sweep are one surface), a
 * heart at its top right; then the brand and riding type in small capitals,
 * the name and the price on one line, the one line a rider compares by, the
 * stock in the owner's words, and an outlined pill to the bike's sizes.
 * Under the pointer the tile shows the drivetrain close-up.
 */
export default function ProductCard({ product }: { product: Product }) {
  const types = useCollection<Department>("categories", { pageSize: 12 }).items;
  const type = types.find((t) => t.id === product.category)?.name;
  const out = (product.stock ?? 0) <= 0;
  return (
    <article data-record={`products:${product.id}`} className="relative flex h-full flex-col">
      <Link to={`/shop/${product.id}`} data-testid={`card-${product.id}`} className="group block">
        <ProductImage src={product.image} hover={product.image_back} alt={`${product.brand} ${product.name}`} aspect="4/3"
          className="rounded border border-line" />
        <span className="mt-4 block text-xs uppercase tracking-[0.14em] text-ink-soft">{product.brand}{type ? ` · ${type}` : ""}</span>
        <span className="mt-1.5 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-0.5">
          <span className="text-base font-semibold text-ink underline-offset-4 group-hover:underline">{product.name}</span>
          <span className="shrink-0 text-base font-bold tabular-nums text-ink">{formatPrice(product.price)}</span>
        </span>
        <span className="mt-1 flex flex-wrap items-baseline justify-between gap-x-4 text-sm text-ink-soft">
          <span>{product.fact}</span>
          {product.was !== undefined && product.was > product.price && (
            <span className="shrink-0 tabular-nums line-through">{formatPrice(product.was)}</span>
          )}
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
