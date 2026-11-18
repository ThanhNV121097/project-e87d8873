import { T, useContent } from "../editable";
import OverviewPage from "../components/admin/OverviewPage";
import AdminLowStock from "../components/AdminLowStock";
import { useCollection } from "../data";
import type { Product } from "../lib/shop";
import { useAdminNav } from "./admin-nav";

/**
 * The first screen of the shop's own side: four figures, each opening the
 * list it counts, and under them what is running low — the question the
 * person signing in at nine is actually asking.
 *
 * Until the records have arrived there is no number anywhere on it — not a
 * figure and not the count inside a note, because "0 still in draft" is a 0
 * too — and the panel does not say "Nothing is running low.", which is a
 * statement about records it has not seen.
 */
export default function AdminOverview() {
  const nav = useAdminNav();
  const bikes = useCollection<Product>("products", { pageSize: 500 });
  const kinds = useCollection("categories", { pageSize: 1 });
  const products = bikes.items;
  const loading = bikes.loading || kinds.loading;
  const low = products.filter((p) => (p.stock ?? 0) < 10);
  const drafts = products.filter((p) => p.status === "draft").length;
  const out = products.filter((p) => (p.stock ?? 0) === 0).length;
  const draftsNote = useContent("pages.admin.overview.drafts");
  const columns = [
    useContent("pages.admin.overview.col_product"),
    useContent("pages.admin.overview.col_kind"),
    useContent("pages.admin.overview.col_stock"),
  ] as const;
  // What stands where a number will be: not a number.
  const figure = (n: number) => (loading ? "—" : n);
  return (
    <OverviewPage
      title={useContent("pages.admin.overview.title")}
      nav={nav}
      stats={[
        { label: useContent("pages.admin.overview.products"), value: figure(products.length), icon: "tag", to: "/admin/products",
          note: `${figure(drafts)} ${draftsNote}` },
        { label: useContent("pages.admin.overview.kinds"), value: figure(kinds.total), icon: "grid", to: "/admin/kinds",
          note: useContent("pages.admin.overview.kinds_note") },
        { label: useContent("pages.admin.overview.low"), value: figure(low.length), icon: "alert", to: "/admin/products",
          note: useContent("pages.admin.overview.low_note") },
        { label: useContent("pages.admin.overview.out"), value: figure(out), icon: "package", to: "/admin/products",
          note: useContent("pages.admin.overview.out_note") },
      ]}
      panels={[{
        title: useContent("pages.admin.overview.low_title"),
        content: loading
          ? <div aria-busy="true" className="h-16" />
          : low.length > 0
            ? <AdminLowStock rows={low} labels={columns} />
            : <T k="pages.admin.overview.low_none" as="p" className="text-sm text-ink-soft" />,
      }]}
    />
  );
}
