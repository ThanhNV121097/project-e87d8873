import type { ReactNode } from "react";

/**
 * A BAND THAT RUNS TO BOTH EDGES OF THE SCREEN while what it holds stays on
 * the page's own edge. The page sits in `max-w-page` with the gutter inside
 * it, so the band reaches out by half the space either side of the measure
 * plus the gutter, and pads back in by the same amount: the heading in a
 * band starts exactly where the heading under it does.
 *
 * Used for the head of a working page — the catalogue, the bag, the
 * checkout — so its first screen is not one more block stopping at the
 * margin, and the colour of the band says "this is where you choose".
 */
export default function Bleed({ children, className = "" }: { children: ReactNode; className?: string }) {
  return (
    <div data-bleed="" className={`-mx-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] px-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] ${className}`}>
      {children}
    </div>
  );
}
