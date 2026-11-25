/**
 * ONE RECORD, OPEN.
 *
 * The other screen of a management area: what the list opens onto. A record
 * with four fields belongs in the list's dialog and never needs this; a
 * record with twenty — a product with its description, its pictures, its
 * price, its stock and its status — needs a page, and the fields group into
 * areas (<Sections>) rather than running down one column for two screenfuls.
 *
 *   export default function Product() {
 *     const { id } = useParams();
 *     return (
 *       <RecordPage collection="products" id={id} nav={ADMIN_NAV}
 *         backTo="/admin/products"
 *         title={useContent("pages.admin.products.title")} />
 *     );
 *   }
 *
 * The fields come from src/data/schema.json; giving a spec a `section` puts
 * it on a tab of that name, and a page that passes none gets one panel.
 *
 * IT READS THE RECORD ITSELF, BY ID (store.ts readRow), because the values
 * on the form and the version token a save is judged by must come from the
 * same answer. Until that answer is in it draws neither the form nor "This
 * record is not here": a record that has not arrived is not a record that is
 * missing. Save changes and Delete wait for the server; a refusal is said in
 * an alert notice (above the form, or inside the delete question) and the
 * typed values stay.
 *
 * WHAT IT REFUSES TO GUESS: a record the server says is not there. An id that
 * is not there is a link somebody followed after the record was deleted, or a
 * typed address, and a page that draws an empty form for it invites the
 * person to save a new record under an id of their own — so it says so and
 * offers the way back.
 */
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import DashboardLayout, { type NavGroup } from "../DashboardLayout";
import { Breadcrumb, Button, Confirm, Empty, Notice, PageHeader, Panel, Sections } from "./index";
import { Form } from "./Form";
import { fieldsOf, missing, titleOf, type FieldSpec } from "./fields";
import { deleteRow, OPEN_FAILED, readRow, saveRow, StoreError, useAttempt } from "./store";
import type { Row } from "../../data";

/** Where the read of the record stands: asked, in hand, not there, or unreadable. */
type Read =
  | { state: "loading" }
  | { state: "ready"; row: Row }
  | { state: "missing" }
  | { state: "failed"; problem: string };

type Values = Record<string, unknown>;

/** The form, in the tabs the fields ask for (in the order they first appear) or one panel. */
function formOf(
  specs: readonly FieldSpec[], values: Values, errors: Record<string, string>,
  change: (key: string, v: unknown) => void,
) {
  const areas: string[] = [];
  for (const f of specs) {
    const area = f.section ?? "";
    if (area && !areas.includes(area)) areas.push(area);
  }
  if (areas.length === 0) return <Form fields={specs} values={values} errors={errors} onChange={change} />;
  return (
    <Sections
      tabs={areas.map((area) => ({
        label: area,
        content: (
          <Form fields={specs.filter((f) => (f.section ?? "") === area)}
            values={values} errors={errors} onChange={change} />
        ),
      }))}
    />
  );
}

export default function RecordPage({
  collection, id, nav, title, brand, fields, backTo, remove = true, children,
}: {
  collection: string;
  /** From useParams(). */
  id: string | undefined;
  nav: readonly NavGroup[];
  /** The AREA's name for the top bar — "Products", not this record's name. */
  title: string;
  brand?: React.ReactNode;
  fields?: readonly FieldSpec[];
  /** The list this record belongs to, for the crumb and after a delete. */
  backTo?: string;
  remove?: boolean;
  /** Under the form: related lists, a log, whatever this record has. */
  children?: React.ReactNode;
}) {
  const navigate = useNavigate();
  const specs = fields ?? fieldsOf(collection);
  const [read, setRead] = useState<Read>({ state: "loading" });
  const [values, setValues] = useState<Values>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [saved, setSaved] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const attempt = useAttempt();
  const item = read.state === "ready" ? read.row : undefined;

  // The record as the server holds it now, and — kept by readRow — the token
  // a save of these values will be judged by.
  useEffect(() => {
    let current = true;
    setRead({ state: "loading" });
    readRow(collection, id ?? "").then(
      (row) => {
        if (!current) return;
        setValues({ ...row });
        setRead({ state: "ready", row });
      },
      (failure: unknown) => {
        if (!current) return;
        const gone = failure instanceof StoreError && failure.status === 404;
        setRead(gone
          ? { state: "missing" }
          : { state: "failed", problem: failure instanceof Error ? failure.message : OPEN_FAILED });
      },
    );
    return () => {
      current = false;
    };
  }, [collection, id]);

  async function save() {
    const bad = missing(specs, values);
    setErrors(bad);
    if (Object.keys(bad).length > 0 || !item) {
      setSaved(false);
      return;
    }
    const done = await attempt.run(async () => {
      const row = await saveRow(collection, { ...values, id: item.id } as Row);
      setValues({ ...row });
      setRead({ state: "ready", row });
    });
    setSaved(done);
  }

  async function destroy() {
    if (!item) return;
    await deleteRow(collection, item.id);
    // Two calls, not navigate(backTo ?? -1): the router's delta overload
    // takes a number and its address overload a string, never a union.
    if (backTo) navigate(backTo);
    else navigate(-1);
  }

  const change = (key: string, v: unknown) => {
    setValues((cur) => ({ ...cur, [key]: v }));
    setSaved(false);
  };

  return (
    <DashboardLayout title={title} nav={nav} brand={brand}>
      {backTo && read.state !== "loading" && (
        <Breadcrumb trail={[{ label: title, to: backTo }, { label: titleOf(item) }]} />
      )}
      {read.state === "missing" && (
        <Empty
          title="This record is not here"
          note="It may have been deleted since the link was made."
          action={backTo ? <Button onClick={() => navigate(backTo)}>Back to the list</Button> : undefined}
        />
      )}
      {read.state === "failed" && <Notice tone="alert">{read.problem}</Notice>}
      {item && (
        <>
          <PageHeader
            title={titleOf(item)}
            actions={
              <>
                {remove && <Button kind="quiet" icon="remove" onClick={() => setConfirming(true)}>Delete</Button>}
                <Button kind="primary" icon="check" onClick={save} disabled={attempt.busy}>Save changes</Button>
              </>
            }
          />
          {saved && <Notice tone="done" onClose={() => setSaved(false)}>Saved.</Notice>}
          {attempt.problem && <Notice tone="alert">{attempt.problem}</Notice>}
          <Panel>{formOf(specs, values, errors, change)}</Panel>
          {children}
          <Confirm
            open={confirming}
            onClose={() => setConfirming(false)}
            onConfirm={destroy}
            title={`Delete ${titleOf(item)}?`}
          />
        </>
      )}
    </DashboardLayout>
  );
}
