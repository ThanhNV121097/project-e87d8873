import type * as React from "react";
import { Picture } from "../picture";

/** The frame's shape. One per grid: every card in a row has the same one. */
export type Aspect = "1/1" | "4/5" | "4/3";

// The size a picture is declared at, so the browser holds its place before it
// loads. The frame, not the file, decides the shape.
const INTRINSIC: Record<Aspect, [number, number]> = {
  "1/1": [800, 800],
  "4/5": [800, 1000],
  "4/3": [800, 600],
};

/** How a product's picture is drawn: see the two kinds below. */
type Kind = "cutout" | "photo";

/** A .png is a picture with no background of its own — an older cut-out, or
 *  a transparent file the owner gave; anything else is a photograph. Known
 *  from the name at once, so nothing is hidden while a file is looked at. */
function kindOf(src: string | null | undefined): Kind | null {
  if (!src) return null;
  return /\.png(?:$|[?#])/i.test(src) ? "cutout" : "photo";
}

const DRAWN: Record<Kind, string> = {
  cutout: "absolute inset-[8%] h-[84%] w-[84%] object-contain",
  photo: "absolute inset-0 h-full w-full object-cover",
};

/**
 * Every product picture on the site: <ProductImage src={p.image} alt={p.name} />.
 *
 * A fixed frame (aspect "1/1" by default, or "4/5", "4/3") — every card in a
 * row has the same one — on a light tile (var(--product-ground) in theme.css)
 * that is seen only until the picture loads.
 *
 * A PRODUCT'S PICTURE IS A PHOTOGRAPH, DRAWN AS IT IS. Nothing here takes a
 * background out, multiplies a white backdrop into the tile or tints it: the
 * owner's decision of 2026-11-16 was that pictures keep their backgrounds.
 * A photograph — the juice with its fruit, the bicycle on its maker's white
 * sweep, the bag carried — FILLS the frame edge to edge, cropped to it. Give
 * every thing in one row the same KIND of photograph (all on white, or all in
 * their place) and let what is in each one differ.
 *   - `.jpg` / `.webp` — a photograph: fills the frame (`object-cover`).
 *   - `.png` — a picture with no background (an older cut-out, a transparent
 *     file the owner supplied): shown whole with an 8% margin on the tile.
 * Never rename a file to move it across.
 *
 * With no `src` the frame shows the flat <Picture> block with the caption;
 * give `slot` (`/images/products/<id>.jpg`) and the owner can drop the photo
 * on it in the Studio. `eager` only for a picture in view on load.
 */
export default function ProductImage({ src, alt, aspect = "1/1", slot, caption, eager = false, className = "", hover }: {
  src?: string | null;
  /** The thing's SECOND picture — its back, a detail, or the thing alone
   *  where the first shows it in use — in place of the first while the
   *  pointer is on its card, on a site at least lively. Each is drawn as the
   *  kind it is. */
  hover?: string | null;
  alt: string;
  aspect?: Aspect;
  slot?: string;
  caption?: string;
  eager?: boolean;
  className?: string;
}) {
  const [width, height] = INTRINSIC[aspect];
  const kind = kindOf(src);
  const over = kindOf(hover);
  return (
    <div data-product-image="" data-product-photo={kind === "photo" || undefined} style={{
      aspectRatio: aspect.replace("/", " / "),
      background: "var(--product-ground)",
      // The tile is light in both modes, so a <Picture> block inside it is
      // drawn on the tile rather than on the site's surface. Without this a
      // product with no photograph yet was a dark hole in a lit grid.
      ["--picture-ground" as string]: "var(--product-ground)",
      ["--picture-ink" as string]: "var(--product-ground-ink)",
    } as React.CSSProperties}
      className={`relative w-full overflow-hidden rounded ${className}`}>
      {src ? (
        <Picture src={src} alt={alt} caption={caption} width={width} height={height} fit={kind === "photo" ? "cover" : "contain"}
          loading={eager ? "eager" : "lazy"}
          className={DRAWN[kind ?? "photo"]} />
      ) : (
        <Picture src={slot ?? ""} alt={alt} caption={caption} fit="contain"
          className="absolute inset-0 h-full w-full" />
      )}
      {src && hover && hover !== src && over && (
        <img data-product-hover="" src={hover} alt="" loading="lazy" decoding="async" className={DRAWN[over]} />
      )}
    </div>
  );
}
