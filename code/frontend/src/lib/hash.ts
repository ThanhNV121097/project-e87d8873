import { useEffect } from "react";
import { useLocation } from "react-router-dom";

/**
 * Brings the element a link's #fragment names into view once the page has
 * drawn it — a search result's workshop job, question or photograph lands on
 * that row, not on the top of the page. A question in a <details> opens.
 *
 * `ready` is whether the records the fragment points into have loaded: they
 * arrive a tick after the page, as they will from the API.
 */
export function useHashScroll(ready: boolean) {
  const { hash } = useLocation();
  useEffect(() => {
    if (!hash || !ready) return;
    const el = document.getElementById(decodeURIComponent(hash.slice(1)));
    if (!el) return;
    if (el instanceof HTMLDetailsElement) el.open = true;
    el.scrollIntoView({ block: "start" });
  }, [hash, ready]);
}
