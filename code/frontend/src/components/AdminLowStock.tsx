import { Link } from "react-router-dom";
import { useContent } from "../editable";
import { useCollection } from "../data";
import { useTableQuery } from "../lib/table";
import type { Department, Product } from "../lib/shop";
import DataTable, { type Column } from "./DataTable";
import TableFilters, { FilterSelect, SearchBox } from "./TableFilters";
import { Badge } from "./admin";
import { nameIn } from "./admin/fields";

/**
 * What is running low, on the overview: the bikes under ten in stock, least
 * first, read through a search and the riding type — both kept in the
 * address bar, so "trail bikes running low" is a link the owner can send —
 * with the table's page numbers and total under it.
 */
export default function AdminLowStock({ rows, labels }: {
  rows: readonly Product[];
  /** The three column headings, from content.json so the owner can change them. */
  labels: readonly [string, string, string];
}) {
  const query = useTableQuery({ size: 20, sort: "stock" });
  const types = useCollection<Department>("categories", { pageSize: 24 }).items;
  const search = useContent("pages.admin.overview.search");
  const any = useContent("pages.admin.overview.any_kind");
  const empty = useContent("pages.admin.overview.low_empty");
  const caption = useContent("pages.admin.overview.low_caption");
  const q = query.q.trim().toLowerCase();
  const kind = query.get("category");
  const matched = rows.filter((p) => (!q || p.name.toLowerCase().includes(q)) && (!kind || p.category === kind));
  const field = query.sort.replace(/^-/, "") || "stock";
  const desc = query.sort.startsWith("-");
  const sorted = [...matched].sort((a, b) => {
    const d = field === "name" ? a.name.localeCompare(b.name) : (a.stock ?? 0) - (b.stock ?? 0);
    return desc ? -d : d;
  });
  const pages = Math.max(1, Math.ceil(sorted.length / query.size));
  const page = Math.min(query.page, pages);
  const shown = sorted.slice((page - 1) * query.size, page * query.size);
  const columns: Column<Product>[] = [
    { key: "name", label: labels[0], sortable: true,
      render: (p) => <Link to={`/admin/products/${p.id}`} className="font-medium text-ink hover:underline">{p.name}</Link> },
    { key: "category", label: labels[1], secondary: true, render: (p) => nameIn("categories", p.category) },
    { key: "stock", label: labels[2], sortable: true, align: "right",
      render: (p) => <Badge tone={(p.stock ?? 0) === 0 ? "danger" : "neutral"}>{p.stock ?? 0}</Badge> },
  ];
  return (
    <div className="flex flex-col gap-4">
      <TableFilters onClear={query.filtering ? query.clear : undefined}>
        <SearchBox value={query.q} onChange={query.setQ} placeholder={search} />
        <FilterSelect label={labels[1]} value={kind} onChange={(v) => query.set("category", v)}
          options={types.map((d) => [d.id, d.name] as const)} anyLabel={any} />
      </TableFilters>
      <DataTable columns={columns} rows={shown} rowKey={(p) => p.id} total={sorted.length} page={page} pages={pages}
        size={query.size} sort={query.sort} onSort={query.toggleSort} onPage={query.setPage} onSize={query.setSize}
        href={(p) => `/admin/products/${p.id}`} caption={caption} exportName="running-low" empty={empty} />
    </div>
  );
}
