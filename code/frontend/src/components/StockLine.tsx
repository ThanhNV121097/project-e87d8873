import { T } from "../editable";
import Icon from "./Icon";

/**
 * Whether the bike can be had, in the owner's own words: "In stock" or
 * "Out of stock" — and, when only a few are left, how few. Always a word
 * with its mark, never a colour alone.
 */
export default function StockLine({ stock, className = "" }: { stock: number; className?: string }) {
  const out = stock <= 0;
  const few = !out && stock <= 3;
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs ${out ? "text-ink-soft" : "text-ink"} ${className}`}>
      <Icon name={out ? "alert" : "done"} size={15} className={out ? "text-danger" : "text-ink"} />
      <T k={out ? "stock.out" : "stock.in"} className="font-medium" />
      {few && (
        <span className="text-ink-soft">· <span className="tabular-nums">{stock}</span> <T k="stock.left" /></span>
      )}
    </span>
  );
}
