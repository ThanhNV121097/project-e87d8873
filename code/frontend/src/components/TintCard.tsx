import { Link } from "react-router-dom";
import { useContent } from "../editable";
import Icon from "./Icon";
import ProductImage from "./ProductImage";
import { formatPrice } from "../lib/format";
import type { Product } from "../lib/shop";

// Written out whole so Tailwind finds every class in the source.
export const TINTS = ["bg-[var(--tint-1)]", "bg-[var(--tint-2)]", "bg-[var(--tint-3)]", "bg-[var(--tint-4)]"] as const;

/**
 * THE TINTED CARD: a tall rounded card filled with ONE pale tint. Its upper
 * part carries the riding type as a small white pill and the model's name
 * large; its lower window is the maker's own studio photograph, edge to edge
 * — a bike is a landscape thing, so the window is 4:3 and nothing is
 * cropped. Across the foot, a bar of glass with the price at the left and a
 * round dark button at the right that opens the bike to choose a size.
 *
 * The template's photographs were shot on the card's colour; no library has
 * a mountain bike on pastel, so here the card carries the colour and the
 * photograph keeps the white sweep it was taken on.
 */
export default function TintCard({ product, tint, label, compact = false, eager = false }: {
  product: Product;
  /** Which of the four tints: the card's place in its row. */
  tint: number;
  /** The small pill at the top left — the riding type. */
  label?: string;
  compact?: boolean;
  eager?: boolean;
}) {
  const choose = useContent("tint.choose");
  return (
    <article data-record={`products:${product.id}`} data-tilt=""
      className={`relative flex flex-col overflow-hidden rounded ${TINTS[tint % TINTS.length]} ${compact ? "" : "min-h-full"}`}>
      <Link to={`/shop/${product.id}`} data-testid={`tint-${product.id}`} className="flex flex-1 flex-col">
        <span className={`flex flex-1 flex-col items-start gap-3 ${compact ? "p-5" : "p-5 pb-8 lg:p-7 lg:pb-10"}`}>
          {label && <span className="rounded-full bg-ground px-3 py-1 text-xs text-ink">{label}</span>}
          <span className={`leading-[1.08] tracking-tight text-ink ${compact ? "text-xl" : "text-[1.75rem] lg:text-[2rem]"}`}>
            <span className="font-light">{product.brand} </span>
            <span className="font-bold">{product.name}</span>
          </span>
          {product.fact && !compact && <span className="text-sm text-ink-soft">{product.fact}</span>}
        </span>
        <ProductImage src={product.image} hover={product.image_back} alt={`${product.brand} ${product.name}`} aspect="4/3"
          eager={eager} className="rounded-none" />
      </Link>
      <div className="absolute inset-x-3 bottom-3 flex items-center justify-between rounded-full border border-line bg-[var(--glass)] py-1.5 pl-5 pr-1.5 backdrop-blur-md">
        <span className="text-sm font-semibold tabular-nums text-ink">{formatPrice(product.price)}</span>
        <Link to={`/shop/${product.id}#sizes`} aria-label={`${choose} ${product.name}`}
          className="grid h-10 w-10 place-items-center rounded-full bg-accent text-accent-ink hover:opacity-90">
          <Icon name="arrow" size={17} />
        </Link>
      </div>
    </article>
  );
}
